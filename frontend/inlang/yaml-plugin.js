// The inlang SDK imports local plugins from a data: URL, where bare specifiers do not resolve,
// so dependencies are resolved from the package directory that every Paraglide compile runs in.
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';

const require = createRequire(`${process.cwd()}/package.json`);
const { parse } = require('yaml');
const { default: messageFormat } = await import(
  pathToFileURL(require.resolve('@inlang/plugin-message-format')).href
);

const key = 'plugin.coolssh.yaml';
const encoder = new TextEncoder();
const decoder = new TextDecoder();

// Catalogs group messages under sections; a message's key joins its path with underscores, as
// internal/catalog does: `keys: { empty: … }` is `keys_empty`.
const flatten = (node, prefix = '') =>
  Object.entries(node ?? {}).reduce((messages, [name, value]) => {
    const path = prefix ? `${prefix}_${name}` : name;
    return typeof value === 'object' && value !== null
      ? Object.assign(messages, flatten(value, path))
      : Object.assign(messages, { [path]: value });
  }, {});

// Reads `{locale}.yaml` catalogs and hands them to the message-format plugin, which owns the
// message syntax.
export default {
  key,
  toBeImportedFiles: async ({ settings }) =>
    settings.locales.map((locale) => ({
      locale,
      path: settings[key].pathPattern.replace('{locale}', locale),
    })),
  importFiles: async ({ files, ...args }) =>
    messageFormat.importFiles({
      ...args,
      files: files.map((file) => ({
        ...file,
        content: encoder.encode(JSON.stringify(flatten(parse(decoder.decode(file.content))))),
      })),
    }),
};
