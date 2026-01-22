const core = require('@actions/core');
const exec = require('@actions/exec');
const io = require('@actions/io');
const tc = require('@actions/tool-cache');
const fs = require('fs');
const os = require('os');
const path = require('path');

const BINARY_NAME = 'bitrise-cache';
const BINARY_TAG = 'v@VERSION@';
const ENVMAN_VERSION = 'latest';

function getPlatform() {
  const platform = os.platform();
  const arch = os.arch();

  let osName;
  let archName;
  let extension = '';

  switch (platform) {
    case 'linux':
      osName = 'Linux';
      break;
    case 'darwin':
      osName = 'Darwin';
      break;
    case 'win32':
      osName = 'Windows';
      extension = '.exe';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  switch (arch) {
    case 'x64':
      archName = 'x86_64';
      break;
    case 'arm64':
      archName = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }

  return {os: osName, arch: archName, extension: extension};
}

async function ensureEnvman(platform) {
  // Check if envman is already available
  try {
    await io.which('envman', true);
    core.debug('envman already available in PATH');
    return;
  } catch {
    core.debug('envman not found in PATH, will install');
  }

  const envmanUrl = `https://github.com/bitrise-io/envman/releases/${ENVMAN_VERSION}/download/envman-${platform.os}-${platform.arch}`;

  core.info(`Installing envman from ${envmanUrl}`);

  await downloadTool('envman', envmanUrl, platform);
}

async function downloadBinary(platform) {
  const url = `https://github.com/bitrise-io/github-cache/releases/download/${BINARY_TAG}/${BINARY_NAME}_${BINARY_TAG}_${platform.os}_${platform.arch}`;

  core.info(`Installing binary from ${url}`);

  return downloadTool(BINARY_NAME, url, platform);
}

async function downloadTool(toolName, url, platform) {
  core.info(`Downloading ${toolName} from ${url}`);

  const from = await tc.downloadTool(url);
  const toPath = path.join(os.homedir(), '.bitrise', 'bin');

  await io.mkdirP(toPath);
  const to = path.join(toPath, `${toolName}${platform.extension}`);
  await io.cp(from, to);
  await fs.promises.chmod(to, 0o755);

  core.addPath(toPath);
  core.debug(`${toolName} installed to ${to}`);

  return to;
}

async function setupEnvstore() {
  const workspace = process.env.GITHUB_WORKSPACE || process.cwd();
  const envstorePath = path.join(workspace, '.envstore.yml');

  if (!fs.existsSync(envstorePath)) {
    await fs.promises.writeFile(envstorePath, '');
    core.debug(`Created envstore at ${envstorePath}`);
  }

  return envstorePath;
}

async function run(phase) {
  try {
    const platform = getPlatform();
    // Ensure envman is available
    await ensureEnvman(platform);

    // Setup envstore
    const envstorePath = await setupEnvstore();

    // Get the binary path
    const binaryPath = await downloadBinary(platform)

    core.debug(`Running ${binaryPath} with phase: ${phase}`);

    // Run the binary with the phase argument
    const exitCode = await exec.exec(binaryPath, [phase], {
      env: {
        ...process.env,
        ENVMAN_ENVSTORE_PATH: envstorePath,
      },
    });

    if (exitCode !== 0) {
      throw new Error(`Binary exited with code ${exitCode}`);
    }
  } catch (error) {
    core.setFailed(error.message);
  }
}

module.exports = {run};
