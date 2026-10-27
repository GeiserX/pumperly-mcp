#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const { install, BIN_PATH } = require("./postinstall.js");

async function main() {
  // npm 12, pnpm and bun skip install scripts by default, so the binary may
  // not be there yet: fetch and verify it now. Logs go to stderr.
  if (!fs.existsSync(BIN_PATH)) {
    await install();
  }
  const env = Object.assign({}, process.env, { TRANSPORT: "stdio" });
  const res = spawnSync(BIN_PATH, process.argv.slice(2), { env, stdio: "inherit" });
  if (res.error) {
    throw res.error;
  }
  process.exit(res.status === null ? 1 : res.status);
}

main().catch((err) => {
  process.stderr.write(`pumperly-mcp: ${err.message}\n`);
  process.exit(1);
});
