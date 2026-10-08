import { IgnorePlugin, type Configuration } from 'webpack';
import { merge } from 'webpack-merge';
import grafanaConfig from './.config/webpack/webpack.config';

const config = async (env: Record<string, string>): Promise<Configuration> => {
  const baseConfig = await grafanaConfig(env);

  // Only plugin entries use AMD; native workers must run without an AMD loader.
  const pluginEntries = baseConfig.entry as Record<string, string>;
  baseConfig.entry = Object.fromEntries(
    Object.entries(pluginEntries).map(([name, entry]) => [
      name,
      { import: entry, library: { type: 'amd' } },
    ])
  );
  baseConfig.output = {
    ...baseConfig.output,
    library: undefined,
    enabledLibraryTypes: ['amd'],
    // Resolve worker chunks relative to their script URL, including Grafana subpaths.
    publicPath: 'auto',
  };

  return merge(baseConfig, {
    // Entry-level AMD no longer implies AMD externals; Grafana still supplies them.
    externalsType: 'amd',
    module: {
      rules: [
        // Keep dependency fonts inside dist; the development default [file]
        // preserves ../node_modules and produces URLs outside the plugin directory.
        {
          test: /\.(woff|woff2|eot|ttf|otf)(\?v=\d+\.\d+\.\d+)?$/,
          type: 'asset/resource',
          generator: {
            filename: 'fonts/[name].[contenthash][ext]',
          },
        },
      ],
    },
    plugins: [
      // GraphiQL's optional WebSocket transport is unused by our HTTP fetcher.
      new IgnorePlugin({ resourceRegExp: /^graphql-ws$/ }),
    ],
  });
};

export default config;
