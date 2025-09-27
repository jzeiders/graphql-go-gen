export interface SchemaSource {
  type?: 'file' | 'url' | 'introspection';
  path?: string;
  url?: string;
  headers?: Record<string, string>;
}

export interface Documents {
  include: string[];
  exclude?: string[];
}

export interface OutputTarget {
  path?: string;
  plugins: string[];
  config?: Record<string, any>;
}

export interface Config {
  schema: SchemaSource[];
  documents: Documents;
  generates: Record<string, OutputTarget>;
  watch?: boolean;
  verbose?: boolean;
  scalars?: Record<string, string>;
}

declare const config: Config;
export default config;