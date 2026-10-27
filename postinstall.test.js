"use strict";

const test = require("node:test");
const assert = require("node:assert");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { expectedSha256, download, getAssetName, install } = require("./postinstall.js");
const VERSION = require("./package.json").version;

const A = "a".repeat(64);
const B = "b".repeat(64);

test("expectedSha256 picks the exact asset line", () => {
  const text = [
    `${A}  pumperly-mcp_${VERSION}_linux_amd64.tar.gz.sbom.json`,
    `${B}  pumperly-mcp_${VERSION}_linux_amd64.tar.gz`,
    "",
  ].join("\n");
  assert.strictEqual(expectedSha256(text, `pumperly-mcp_${VERSION}_linux_amd64.tar.gz`), B);
});

test("expectedSha256 accepts CRLF lines and the binary-mode marker", () => {
  const text = `${A} *other.tar.gz\r\n${B} *pumperly-mcp.tar.gz\r\n`;
  assert.strictEqual(expectedSha256(text, "pumperly-mcp.tar.gz"), B);
  assert.throws(() => expectedSha256(`${B}pumperly-mcp.tar.gz\n`, "pumperly-mcp.tar.gz"), /no checksum/);
});

test("expectedSha256 stays linear on a line of many spaces", () => {
  // A backtracking regex takes quadratic time on this line (CodeQL js/polynomial-redos).
  const text = `${A}${" ".repeat(200_000)}x\u2028x\n${B}  pumperly-mcp.tar.gz\n`;
  const start = process.hrtime.bigint();
  assert.strictEqual(expectedSha256(text, "pumperly-mcp.tar.gz"), B);
  const ms = Number(process.hrtime.bigint() - start) / 1e6;
  assert.ok(ms < 1000, `took ${ms} ms`);
});

test("expectedSha256 throws when the asset is not listed", () => {
  assert.throws(() => expectedSha256(`${A}  other.tar.gz\n`, "pumperly-mcp.tar.gz"), /no checksum/);
});

test("getAssetName maps node platform names to goreleaser ones", () => {
  assert.strictEqual(getAssetName("win32", "x64"), `pumperly-mcp_${VERSION}_windows_amd64.zip`);
  assert.strictEqual(getAssetName("darwin", "arm64"), `pumperly-mcp_${VERSION}_darwin_arm64.tar.gz`);
  assert.throws(() => getAssetName("aix", "ppc64"), /Unsupported platform/);
});

test("download refuses plain http before any request", async () => {
  const dest = path.join(fs.mkdtempSync(path.join(os.tmpdir(), "pmcp-test-")), "x");
  await assert.rejects(download("http://example.com/x.tar.gz", dest), /non-https/);
  assert.strictEqual(fs.existsSync(dest), false);
});

test("install refuses to run without the bundled checksums.txt", async (t) => {
  if (fs.existsSync(path.join(__dirname, "checksums.txt")) || fs.existsSync(path.join(__dirname, "bin"))) {
    t.skip("a packaged checkout; nothing to prove here");
    return;
  }
  await assert.rejects(install(), /checksums\.txt is missing/);
});
