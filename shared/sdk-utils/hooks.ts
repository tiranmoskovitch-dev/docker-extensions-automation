/**
 * Shared React hooks for Docker Extensions SDK.
 * Import these in any extension's frontend.
 */

import { useEffect, useState, useCallback } from "react";
import { createDockerDesktopClient } from "@docker/extension-api-client";

const ddClient = createDockerDesktopClient();

export { ddClient };

/**
 * Hook to list Docker containers with auto-refresh.
 */
export function useContainers(refreshInterval = 5000) {
  const [containers, setContainers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const result = await ddClient.docker.listContainers({ all: true });
      setContainers(result as any[]);
      setError(null);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, refreshInterval);
    return () => clearInterval(interval);
  }, [refresh, refreshInterval]);

  return { containers, loading, error, refresh };
}

/**
 * Hook to call the extension backend API.
 */
export function useBackendAPI<T>(path: string) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    try {
      const result = await ddClient.extension.vm?.service?.get(path);
      setData(result as T);
      setError(null);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }, [path]);

  const post = useCallback(async (body: any) => {
    try {
      const result = await ddClient.extension.vm?.service?.post(path, body);
      return result;
    } catch (err) {
      setError(String(err));
      throw err;
    }
  }, [path]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refresh: fetch, post };
}

/**
 * Hook to stream Docker events.
 */
export function useDockerEvents(onEvent: (event: any) => void) {
  useEffect(() => {
    const exec = ddClient.docker.cli.exec("events", ["--format", "json"], {
      stream: {
        onOutput(data: { stdout?: string }) {
          if (data.stdout) {
            try {
              onEvent(JSON.parse(data.stdout));
            } catch {
              // ignore parse errors
            }
          }
        },
        onClose() {},
        onError(err: any) {
          console.error("Docker events stream error:", err);
        },
      },
    });

    return () => {
      // Cleanup: the exec promise handles stream termination
    };
  }, [onEvent]);
}

/**
 * Show a toast notification.
 */
export const toast = {
  success: (msg: string) => ddClient.desktopUI.toast.success(msg),
  warning: (msg: string) => ddClient.desktopUI.toast.warning(msg),
  error: (msg: string) => ddClient.desktopUI.toast.error(msg),
};

/**
 * Open a URL in the system browser.
 */
export const openUrl = (url: string) => ddClient.host.openUrl(url);
