export type PackageManager = 'bun' | 'pnpm';

export type LogLevel = 'info' | 'success' | 'warning' | 'error' | 'spinner';

export interface LogEntry {
  level: LogLevel;
  message: string;
}

export interface ProjectConfig {
  database: string;
  features: {
    stripe: boolean;
    storage: boolean;
    email: boolean;
  };
  name: string;
  projectName: string;
}

export interface State {
  features: string[];
  generated: string[];
}

export interface GenerateContext {
  config: ProjectConfig;
  hasOexctl: boolean;
  packageManager: PackageManager;
  state: State;
  targetDir: string;
}
