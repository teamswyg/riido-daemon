#!/usr/bin/env node
// Real Apple signature verification and real installer; temporary roots only.
// Usage: node scripts/verify-desktop-package-upgrade.cjs DESKTOP_ROOT NEW_PACKAGE OLD_HELPER_PACKAGE_DIR
const assert = require('node:assert/strict');
const fs = require('node:fs');
const fsp = require('node:fs/promises');
const path = require('node:path');
const os = require('node:os');
const vm = require('node:vm');
const { execFileSync } = require('node:child_process');
const [desktop, archive, previous] = process.argv.slice(2).map(p => path.resolve(p));
const ts = require(path.join(desktop, 'node_modules/typescript'));
const cache = new Map();
function load(name) {
  if (cache.has(name)) return cache.get(name);
  // Only these installer modules are needed; no desktop main/daemon startup.
  assert.ok(['daemonCLIInstall', 'daemonCLIHelperSignature', 'daemonMachO'].includes(name));
  const filename = path.join(desktop, 'src/modules', `${name}.ts`);
  const source = ts.transpileModule(fs.readFileSync(filename, 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const exports = {};
  cache.set(name, exports);
  vm.runInNewContext(source, {
    exports, Buffer, process,
    require: id => id.startsWith('./') ? load(id.slice(2)) : require(id),
  }, { filename });
  return exports;
}
(async () => {
  assert.equal(process.platform, 'darwin');
  const root = await fsp.mkdtemp(path.join(os.tmpdir(), 'riido-real-upgrade-'));
  try {
    const source = path.join(root, 'package');
    await fsp.mkdir(source);
    // Release validator must run before this smoke test on the supplied archive.
    execFileSync('/usr/bin/tar', ['-xzf', archive, '-C', source]);
    const version = (await fsp.readFile(path.join(source, 'VERSION'), 'utf8')).trim();
    const oldVersion = (await fsp.readFile(path.join(previous, 'VERSION'), 'utf8')).trim();
    assert.notEqual(version, oldVersion, 'test must cross a release version');
    const installer = load('daemonCLIInstall');
    const checks = [];
    for (const kind of ['flat', 'managed']) {
      const binaryPath = path.join(root, kind, 'bin/riido');
      await fsp.mkdir(path.dirname(binaryPath), { recursive: true });
      if (kind === 'managed') {
        await installer.installCLIAuthBundle({
          binaryPath, extractedBinaryPath: path.join(previous, 'riido'),
          version: oldVersion, etag: 'old',
        });
        assert.equal((await installer.inspectCLIInstall(binaryPath)).healthy, true);
      } else {
        await fsp.copyFile(path.join(previous, 'riido'), binaryPath);
        await fsp.chmod(binaryPath, 0o755);
        assert.equal(await installer.inspectCLIInstall(binaryPath), null);
      }
      const oldTarget = await fsp.realpath(binaryPath);
      const oldHash = require('node:crypto').createHash('sha256').update(await fsp.readFile(oldTarget)).digest('hex');
      let activated = 0;
      await installer.installCLIAuthBundle({
        binaryPath, extractedBinaryPath: path.join(source, 'riido'), version, etag: 'new',
        beforeActivate: async () => {
          activated++;
          assert.equal(await fsp.realpath(binaryPath), oldTarget);
          assert.equal(require('node:crypto').createHash('sha256').update(await fsp.readFile(oldTarget)).digest('hex'), oldHash);
        },
      });
      assert.equal(activated, 1);
      const state = await installer.inspectCLIInstall(binaryPath);
      assert.equal(state.healthy, true);
      assert.equal(state.version, version);
      assert.equal(state.etag, 'new');
      const target = await fsp.realpath(binaryPath);
      assert.notEqual(target, oldTarget);
      assert.ok(execFileSync(binaryPath, ['version'], { encoding: 'utf8' }).includes(version));
      execFileSync('/usr/bin/xcrun', ['stapler', 'validate', path.join(path.dirname(target), 'RiidoCLIAuthHelper.app')]);
      checks.push({ installation: kind, previous: oldVersion, version, signature: 'verified', notarization: 'verified', status: 'pass' });
    }
    console.log(JSON.stringify({ checks }, null, 2));
  } finally { await fsp.rm(root, { recursive: true, force: true }); }
})().catch(error => { console.error(error); process.exitCode = 1; });
