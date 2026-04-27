import type { Configuration } from 'webpack';
import { merge } from 'webpack-merge';
import grafanaConfig from './.config/webpack/webpack.config';

const config = async (env: Record<string, string>): Promise<Configuration> => {
  const baseConfig = await grafanaConfig(env);

  return merge(baseConfig, {
    module: {
      rules: [
        // graphiql v5 (and @graphiql/react) declare sideEffects without listing
        // CSS files, causing webpack to tree-shake their CSS imports. This rule
        // forces all CSS modules to be treated as having side effects.
        // Fixed in https://github.com/graphql/graphiql/pull/4211 but no release yet.
        {
          test: /\.css$/,
          sideEffects: true,
        },
      ],
    },
  });
};

export default config;
