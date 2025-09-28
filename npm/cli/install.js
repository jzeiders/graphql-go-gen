#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const platform = process.platform; // 'darwin', 'linux', 'win32'
const arch = process.arch; // 'arm64', 'x64', ...

const vendorDir = path.join(__dirname, 'vendor');
const exeName = platform === 'win32' ? 'graphql-go-gen.exe' : 'graphql-go-gen';
const vendorBin = path.join(vendorDir, exeName);

const pkgMap = {
  'darwin-arm64': '@graphql-go-gen/cli-darwin-arm64'
};

fs.mkdirSync(vendorDir, { recursive: true });

function installFromOptionalDependency(pkgName) {
  if (!pkgName) {
    return false;
  }

  try {
    const pkgJsonPath = require.resolve(`${pkgName}/package.json`, { paths: [__dirname] });
    const pkgRoot = path.dirname(pkgJsonPath);
    const src = path.join(pkgRoot, 'bin', exeName);
    const data = fs.readFileSync(src);
    fs.writeFileSync(vendorBin, data, { mode: 0o755 });
    fs.chmodSync(vendorBin, 0o755);
    return true;
  } catch (error) {
    if (process.env.DEBUG) {
      console.warn(`[graphql-go-gen] Unable to copy optional dependency ${pkgName}:`, error);
    }
    return false;
  }
}

(() => {
  const key = `${platform}-${arch}`;
  const pkgName = pkgMap[key];

  if (installFromOptionalDependency(pkgName)) {
    return;
  }

  console.error(`[graphql-go-gen] No prebuilt binary available for ${platform}/${arch}.`);
  console.error('[graphql-go-gen] Currently supported platform: darwin arm64.');
  process.exit(1);
})();
