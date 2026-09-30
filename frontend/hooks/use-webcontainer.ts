import { FileSystemTree, WebContainer } from "@webcontainer/api";
import React, { useEffect, useState } from "react";

export interface wcInterface {
  files: FileSystemTree;
  iframeRef: React.RefObject<HTMLIFrameElement | null>;
}

let webContainerPromise: Promise<WebContainer> | null = null;

function getWebContainer() {
  if (!webContainerPromise) {
    webContainerPromise = WebContainer.boot();
  }

  return webContainerPromise;
}

export function useWebContainer({ files, iframeRef }: wcInterface) {
  const [wcInstance, setWcInstance] = useState<WebContainer | null>(null);
  const [status, setStatus] = useState("booting");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function bootWC() {
      try {
        setStatus("booting");

        const webcontainerInstance = await getWebContainer();

        if (cancelled) return;

        setWcInstance(webcontainerInstance);
        setStatus("mounting");
        await webcontainerInstance.mount(files);

        const installProcess = await webcontainerInstance.spawn("npm", [
          "install",
        ]);

        const installExitCode = await installProcess.exit;

        if (installExitCode !== 0) {
          throw new Error("Unable to run npm install");
        }

        if (cancelled) return;

        setStatus("starting");
        webcontainerInstance.on("server-ready", (port, url) => {
          console.log("Server ready:", url);
          if (cancelled) return;
          if (iframeRef.current) {
            iframeRef.current.src = url;
          }
          setStatus("ready");
        });

        await webcontainerInstance.spawn("npm", ["run", "dev"]);
      } catch (e) {
        console.error("WebContainer error:", e);

        if (!cancelled) {
          setError(
            e instanceof Error ? e.message : "Failed to start WebContainer",
          );
          setStatus("error");
        }
      }
    }

    bootWC();

    return () => {
      cancelled = true;
    };
  }, [files, iframeRef]);

  return {
    wcInstance,
    status,
    error,
  };
}
