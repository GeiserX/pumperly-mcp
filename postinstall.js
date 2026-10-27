#!/usr/bin/env node
"use strict";

// Downloads the release binary for this platform and checks it against the
// sha256 in checksums.txt, which the release workflow puts INSIDE this npm
// package. The npm package (published with provenance) is therefore the root
// of trust; nothing fetched from the mutable GitHub release can change what
// is accepted.
//
// Runs as the postinstall script, and from run.js on first use when the
// package manager skipped install scripts (npm 12 does by default). It logs to
// stderr only: when run.js calls it, stdout is the MCP channel.

const { execFileSync } = require("child_process");
const crypto = require("crypto");
const fs = require("fs");
const os = require("os");
const path = require("path");
const https = require("https");

const VERSION = require("./package.json").version;
const REPO = "GeiserX/pumperly-mcp";
const BIN_NAME = process.platform === "win32" ? "pumperly-mcp.exe" : "pumperly-mcp";
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, BIN_NAME);
const CHECKSUMS_PATH = path.join(__dirname, "checksums.txt");

const MAX_REDIRECTS = 10;
const REQUEST_TIMEOUT_MS = 30_000;
// The archives are about 5 MB; anything far larger is not ours.
const MAX_DOWNLOAD_BYTES = 64 * 1024 * 1024;

function log(msg) {
  process.stderr.write(`pumperly-mcp: ${msg}\n`);
}

function getPlatformArch(platform = process.platform, arch = process.arch) {
  const platformMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };
  const p = platformMap[platform];
  const a = archMap[arch];
  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  return { platform: p, arch: a };
}

function getAssetName(platform, arch) {
  const pa = getPlatformArch(platform, arch);
  const ext = pa.platform === "windows" ? "zip" : "tar.gz";
  return `pumperly-mcp_${VERSION}_${pa.platform}_${pa.arch}.${ext}`;
}

// expectedSha256 returns the hash listed for assetName in a goreleaser
// checksums.txt ("<sha256>  <file>" per line), or throws.
function expectedSha256(checksumsText, assetName) {
  for (const line of checksumsText.split(/\r?\n/)) {
    const m = line.trim().match(/^([0-9a-f]{64})\s+\*?(.+)$/);
    if (m && m[2] === assetName) {
      return m[1];
    }
  }
  throw new Error(`no checksum for ${assetName} in the bundled checksums.txt`);
}

// download streams url to dest over https only, following at most
// MAX_REDIRECTS https redirects, capping the size, and resolves to the
// sha256 of what was written.
function download(url, dest, redirects = 0) {
  return new Promise((resolve, reject) => {
    let parsed;
    try {
      parsed = new URL(url);
    } catch {
      return reject(new Error("invalid download URL"));
    }
    if (parsed.protocol !== "https:") {
      return reject(new Error(`refusing a non-https download URL (${parsed.protocol})`));
    }
    if (redirects > MAX_REDIRECTS) {
      return reject(new Error(`too many redirects (>${MAX_REDIRECTS})`));
    }
    const req = https.get(parsed, { timeout: REQUEST_TIMEOUT_MS }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        const next = new URL(res.headers.location, parsed).toString();
        return download(next, dest, redirects + 1).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`download failed: HTTP ${res.statusCode}`));
      }
      const hash = crypto.createHash("sha256");
      const out = fs.createWriteStream(dest, { mode: 0o600 });
      let size = 0;
      res.on("data", (chunk) => {
        size += chunk.length;
        if (size > MAX_DOWNLOAD_BYTES) {
          req.destroy(new Error(`download larger than ${MAX_DOWNLOAD_BYTES} bytes`));
          return;
        }
        hash.update(chunk);
      });
      res.on("error", reject);
      out.on("error", reject);
      res.pipe(out);
      out.on("finish", () => {
        if (size > MAX_DOWNLOAD_BYTES) return;
        resolve(hash.digest("hex"));
      });
    });
    req.on("timeout", () => req.destroy(new Error(`request timed out after ${REQUEST_TIMEOUT_MS} ms`)));
    req.on("error", reject);
  });
}

// extract unpacks only the binary from the archive into dir, without a shell.
function extract(archive, dir) {
  if (archive.endsWith(".zip")) {
    if (process.platform === "win32") {
      execFileSync(
        "powershell",
        [
          "-NoProfile",
          "-NonInteractive",
          "-Command",
          "Expand-Archive -Force -LiteralPath $env:PMCP_ARCHIVE -DestinationPath $env:PMCP_DEST",
        ],
        { stdio: "ignore", env: { ...process.env, PMCP_ARCHIVE: archive, PMCP_DEST: dir } }
      );
    } else {
      execFileSync("unzip", ["-o", archive, BIN_NAME, "-d", dir], { stdio: "ignore" });
    }
  } else {
    execFileSync("tar", ["-xzf", archive, "-C", dir, BIN_NAME], { stdio: "ignore" });
  }
}

async function install() {
  if (fs.existsSync(BIN_PATH)) {
    return BIN_PATH;
  }
  if (!fs.existsSync(CHECKSUMS_PATH)) {
    throw new Error(
      "checksums.txt is missing from the package, so the download cannot be verified. Install a published release from npm."
    );
  }
  const assetName = getAssetName();
  const want = expectedSha256(fs.readFileSync(CHECKSUMS_PATH, "utf8"), assetName);
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${assetName}`;

  fs.mkdirSync(BIN_DIR, { recursive: true });
  const work = fs.mkdtempSync(path.join(os.tmpdir(), "pumperly-mcp-"));
  try {
    const archive = path.join(work, assetName);
    log(`downloading v${VERSION} for ${process.platform}-${process.arch}...`);
    const got = await download(url, archive);
    if (got !== want) {
      throw new Error(`checksum mismatch for ${assetName}: expected ${want}, got ${got}`);
    }
    extract(archive, work);
    const extracted = path.join(work, BIN_NAME);
    if (process.platform !== "win32") {
      fs.chmodSync(extracted, 0o755);
    }
    // Copy next to the final path, then rename: a concurrent first run never
    // sees a half-written binary.
    const staged = `${BIN_PATH}.${process.pid}.tmp`;
    fs.copyFileSync(extracted, staged);
    if (process.platform !== "win32") {
      fs.chmodSync(staged, 0o755);
    }
    fs.renameSync(staged, BIN_PATH);
    log(`installed ${BIN_PATH} (sha256 verified)`);
    return BIN_PATH;
  } finally {
    fs.rmSync(work, { recursive: true, force: true });
  }
}

module.exports = { install, expectedSha256, download, getAssetName, BIN_PATH };

if (require.main === module) {
  install().catch((err) => {
    log(`install failed: ${err.message}`);
    process.exit(1);
  });
}
