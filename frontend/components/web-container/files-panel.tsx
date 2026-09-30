import { Tree, type TreeViewElement } from "@/components/ui/file-tree";
import type { FileSystemTree } from "@webcontainer/api";

export function createTreeElements(files: FileSystemTree) {
  const expandedItems: string[] = [];
  let firstFileId: string | undefined;

  function visitTree(
    entries: FileSystemTree,
    parentPath = "",
  ): TreeViewElement[] {
    return Object.entries(entries).map(([name, entry]) => {
      const path = parentPath ? `${parentPath}/${name}` : name;

      if ("directory" in entry) {
        expandedItems.push(path);
        return {
          id: path,
          type: "folder",
          isSelectable: true,
          name,
          children: visitTree(entry.directory, path),
        };
      }

      firstFileId ??= path;
      return {
        id: path,
        isSelectable: true,
        name,
      };
    });
  }

  return {
    elements: visitTree(files),
    expandedItems,
    firstFileId,
  };
}

export function FilesPanel({
  files,
}: {
  files: ReturnType<typeof createTreeElements>;
}) {
  return (
    <Tree
      className="min-h-0 flex-1 overflow-auto p-2"
      initialSelectedId={files.firstFileId}
      initialExpandedItems={files.expandedItems}
      elements={files.elements}
    />
  );
}
