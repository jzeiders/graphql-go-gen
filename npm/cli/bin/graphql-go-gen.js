#!/usr/bin/env node
const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

const isWin = process.platform === 'win32';
const exeName = isWin ? 'graphql-go-gen.exe' : 'graphql-go-gen';
const binPath = path.join(__dirname, '..', 'vendor', exeName);

if (!fs.existsSync(binPath)) {
  console.error('[graphql-go-gen] Binary is missing. Try reinstalling the package.');
  process.exit(1);
}

const child = spawn(binPath, process.argv.slice(2), {
  stdio: 'inherit'
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 0);
  }
});

child.on('error', (error) => {
  console.error('[graphql-go-gen] Failed to launch binary:', error.message);
  process.exit(1);
});
