// Wrappers around the native OS pickers exposed by DialogBindings.
// We hit window.go directly so the call works regardless of whether
// the generated .d.ts is stale (wails dev only regenerates on cold start).

type DialogBindingsApi = {
  SelectFile: (title: string) => Promise<string>;
  SelectFolder: (title: string) => Promise<string>;
  SaveFile: (defaultName: string, title: string) => Promise<string>;
  SaveArchive: (defaultName: string) => Promise<string>;
};

function bindings(): DialogBindingsApi {
  const ns = (window as unknown as { go?: { wailsapp?: { DialogBindings?: DialogBindingsApi } } })
    .go?.wailsapp?.DialogBindings;
  if (!ns) {
    throw new Error("DialogBindings not available on window.go");
  }
  return ns;
}

export const dialogsApi = {
  selectFile: (title = ""): Promise<string> => bindings().SelectFile(title),
  selectFolder: (title = ""): Promise<string> => bindings().SelectFolder(title),
  saveFile: (defaultName = "", title = ""): Promise<string> =>
    bindings().SaveFile(defaultName, title),
  saveArchive: (defaultName = ""): Promise<string> =>
    bindings().SaveArchive(defaultName),
};
