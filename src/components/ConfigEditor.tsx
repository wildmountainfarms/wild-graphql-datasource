import React from 'react';
import { css } from '@emotion/css';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import {
  AdvancedHttpSettings,
  Auth,
  ConfigSection,
  ConnectionSettings,
  convertLegacyAuthProps,
} from '@grafana/plugin-ui';
import { WildGraphQLDataSourceOptions } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<WildGraphQLDataSourceOptions> {}

// Match plugin-ui's connection, authentication, and advanced HTTP field widths.
const advancedSettingsClass = css({ maxWidth: 578 });

export function ConfigEditor({ onOptionsChange, options }: Props) {
  return (
    <div className="gf-form-group">
      <ConnectionSettings
        config={options}
        onChange={onOptionsChange}
        urlPlaceholder="http://localhost:8080"
      />
      <Auth {...convertLegacyAuthProps({ config: options, onChange: onOptionsChange })} />
      <ConfigSection
        title="Advanced settings"
        className={advancedSettingsClass}
        isCollapsible
        isInitiallyOpen={Boolean(options.jsonData.keepCookies?.length || options.jsonData.timeout !== undefined)}
      >
        <AdvancedHttpSettings config={options} onChange={onOptionsChange} />
      </ConfigSection>
    </div>
  );
}
