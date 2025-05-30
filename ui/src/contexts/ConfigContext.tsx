import { createContext, useContext } from 'react';

export type Config = {
  apiURL: string;
  basePath: string;
  title: string;
  navbarColor: string;
  tz: string;
  tzOffsetInSec: number | undefined;
  version: string;
  maxDashboardPageLimit: number;
  remoteNodes: string;
  permissions: {
    writeDags: boolean;
    runDags: boolean;
  };
};

export const ConfigContext = createContext<Config>(null!);

export function useConfig() {
  return useContext(ConfigContext);
}
