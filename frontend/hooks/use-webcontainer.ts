import { FileSystemTree, WebContainer } from "@webcontainer/api";
import React, { useCallback, useEffect, useState } from "react";

export interface wcInterface {
  files: FileSystemTree;
  iframeRef: React.RefObject<HTMLIFrameElement | null>;
}

type WCProcess = Awaited<ReturnType<WebContainer["spawn"]>>;

let webContainerPromise: Promise<WebContainer> | null = null;

const ESC = "(?:\\u001b|\\uFFFD)";

const SPINNER_RE = new RegExp(
  `(?:${ESC}\\[[0-9;?]*G${ESC}\\[[0-9;?]*K[\\\\|/-]?)+`,
  "g",
);
const OSC_RE = /\u001b\][\s\S]*?(?:\u0007|\u001b\\)/g;
const CSI_RE = /\u001b\[[0-?]*[ -/]*[@-~]/g;

const INCOMPLETE_TAIL_RE = /(?:\u001b|\uFFFD)(?:\[[0-?]*[ -/]*)?$/;

export function createTerminalCleaner() {
  let pending = "";

  return function clean(chunk: string): string {
    let data = pending + chunk;
    pending = "";

    const tail = data.match(INCOMPLETE_TAIL_RE);
    if (tail) {
      pending = tail[0];
      data = data.slice(0, tail.index);
    }

    return data
      .replace(SPINNER_RE, "")
      .replace(OSC_RE, "")
      .replace(CSI_RE, "")
      .replace(/\r(?!\n)/g, "")
      .replace(/\r\n/g, "\n")
      .split("\n")
      .filter((line) => !/^[\\|/-]$/.test(line.trim()))
      .join("\n")
      .replace(/\n{3,}/g, "\n\n");
  };
}

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
  const [terminalOutput, setTerminalOutput] = useState("");
  const [isCommandRunning, setIsCommandRunning] = useState(false);

  const readFile = useCallback(async (path: string) => {
    const webcontainerInstance = await getWebContainer();
    return webcontainerInstance.fs.readFile(path, "utf-8");
  }, []);

  const writeFile = useCallback(async (path: string, contents: string) => {
    const webcontainerInstance = await getWebContainer();
    await webcontainerInstance.fs.writeFile(path, contents);
  }, []);

  const appendTerminalOutput = useCallback((text: string) => {
    if (!text) return;
    setTerminalOutput((current) => `${current}${text}`.slice(-100_000));
  }, []);

  const pipeOutput = useCallback(
    (process: WCProcess, label: string) => {
      const clean = createTerminalCleaner();
      void process.output
        .pipeTo(
          new WritableStream({
            write(data) {
              appendTerminalOutput(clean(data));
            },
          }),
        )
        .catch((streamError) => {
          appendTerminalOutput(
            `\n${label} output error: ${streamError instanceof Error ? streamError.message : String(streamError)}\n`,
          );
        });
    },
    [appendTerminalOutput],
  );

  const streamProcess = useCallback(
    async function streamProcess(process: WCProcess, command: string) {
      appendTerminalOutput(`\n$ ${command}\n`);
      pipeOutput(process, command);

      const exitCode = await process.exit;
      appendTerminalOutput(`\n[process exited with code ${exitCode}]\n`);
      return exitCode;
    },
    [appendTerminalOutput, pipeOutput],
  );

  const runCommand = useCallback(
    async (commandLine: string) => {
      const webcontainerInstance = await getWebContainer();
      const [command, ...args] = commandLine.trim().split(/\s+/);
      if (!command) return;

      setIsCommandRunning(true);
      try {
        const process = await webcontainerInstance.spawn(command, args);
        await streamProcess(process, commandLine.trim());
      } catch (commandError) {
        appendTerminalOutput(
          `\nCommand error: ${commandError instanceof Error ? commandError.message : String(commandError)}\n`,
        );
      } finally {
        setIsCommandRunning(false);
      }
    },
    [appendTerminalOutput, streamProcess],
  );

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
        const installExitCode = await streamProcess(
          installProcess,
          "npm install",
        );

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

        const devProcess = await webcontainerInstance.spawn("npm", [
          "run",
          "dev",
        ]);
        appendTerminalOutput("\n$ npm run dev\n");
        pipeOutput(devProcess, "Dev server");
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
  }, [appendTerminalOutput, files, iframeRef, streamProcess, pipeOutput]);

  return {
    wcInstance,
    status,
    error,
    terminalOutput,
    isCommandRunning,
    runCommand,
    readFile,
    writeFile,
  };
}
