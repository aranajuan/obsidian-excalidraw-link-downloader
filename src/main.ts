import {
  App,
  normalizePath,
  Notice,
  Plugin,
  PluginSettingTab,
  Setting,
  TFile,
} from 'obsidian';


import { parseAll } from './link';
import { download, ErrEmptyCanvas } from './downloader';
import {
  parseAttachmentFolderPath,
  resolveDestDir,
  buildAnnotation,
  hasExistingLocalCopy,
  injectAll,
  ANNOTATION_WARNING,
  ANNOTATION_REFRESH_FAILED,
  ANNOTATION_EMPTY_CANVAS,
} from './injector';

interface Settings {
  /** Override the vault's attachmentFolderPath. Empty = read from vault config. */
  attachmentFolderOverride: string;
}

const DEFAULT_SETTINGS: Settings = {
  attachmentFolderOverride: '',
};

export default class ExcalidrawDownloaderPlugin extends Plugin {
  settings!: Settings;
  private abortController = new AbortController();
  private progressNotice: Notice | undefined;

  async onload() {
    await this.loadSettings();

    this.addCommand({
      id: 'download-current-file',
      name: 'Download Excalidraw links in current file',
      editorCallback: async (_editor, ctx) => {
        if (!ctx.file) { new Notice('No active file.'); return; }
        await this.processFile(ctx.file, false);
      },
    });

    this.addCommand({
      id: 'refresh-current-file',
      name: 'Refresh Excalidraw local copies in current file',
      editorCallback: async (_editor, ctx) => {
        if (!ctx.file) { new Notice('No active file.'); return; }
        await this.processFile(ctx.file, true);
      },
    });

    this.addSettingTab(new ExcalidrawDownloaderSettingTab(this.app, this));
  }

  onunload() {
    this.progressNotice?.hide();
    this.progressNotice = undefined;
    this.abortController.abort();
  }

  // ── Core processing ────────────────────────────────────────────────────────

  private attachmentFolderPath(): string {
    if (this.settings.attachmentFolderOverride) {
      return normalizePath(this.settings.attachmentFolderOverride);
    }
    return this.app.getConfig('attachmentFolderPath') ?? '';
  }

  async processFile(file: TFile, refresh: boolean): Promise<void> {
    const signal = this.abortController.signal;
    const content = await this.app.vault.read(file);
    const links = parseAll(content, refresh);

    if (links.length === 0) {
      new Notice(`No ${refresh ? 'refreshable' : 'new'} Excalidraw links in ${file.basename}.`);
      return;
    }

    const cfg = parseAttachmentFolderPath(this.attachmentFolderPath());
    const noteDir = file.parent?.path ?? '';
    const destDir = resolveDestDir(cfg, noteDir);

    // In refresh mode, record which URLs already had a local copy so we can
    // preserve their annotation if the re-download fails.
    const hadLocalCopy = new Map<string, boolean>();
    if (refresh) {
      for (const link of links) {
        hadLocalCopy.set(link.url, hasExistingLocalCopy(content, link.url));
      }
    }

    const annotations = new Map<string, string>();
    let downloaded = 0, cached = 0, empty = 0, errors = 0, refreshFailed = 0;

    for (const link of links) {
      const notify = (msg: string) => {
        this.progressNotice?.hide();
        this.progressNotice = new Notice(msg, 10_000); // 10s auto-dismiss
      };

      try {
        const result = await download(link, destDir, this.app, refresh, notify, signal);
        this.progressNotice?.hide();
        const filename = result.destPath.split('/').pop()!;
        annotations.set(link.url, buildAnnotation(filename));
        result.cached ? cached++ : downloaded++;
      } catch (err) {
        this.progressNotice?.hide();
        if (err === ErrEmptyCanvas) {
          annotations.set(link.url, ANNOTATION_EMPTY_CANVAS);
          empty++;
        } else if (refresh && hadLocalCopy.get(link.url)) {
          annotations.set(link.url, ANNOTATION_REFRESH_FAILED);
          refreshFailed++;
          console.warn(`[Excalidraw Downloader] refresh failed for ${link.id}:`, err);
          new Notice(`Excalidraw: refresh failed for link — see console for details`, 5_000);
        } else {
          annotations.set(link.url, ANNOTATION_WARNING);
          errors++;
          console.error(`[Excalidraw Downloader] download failed for ${link.id}:`, err);
          new Notice(`Excalidraw: download failed for link — see console for details`, 5_000);
        }
      } finally {
        this.progressNotice?.hide();
      }
    }

    const newContent = injectAll(content, annotations, refresh);
    if (newContent !== content) {
      await this.app.vault.modify(file, newContent);
    }

    const parts: string[] = [];
    if (downloaded)    parts.push(`${downloaded} downloaded`);
    if (cached)        parts.push(`${cached} cached`);
    if (empty)         parts.push(`${empty} empty canvas`);
    if (errors)        parts.push(`${errors} failed`);
    if (refreshFailed) parts.push(`${refreshFailed} refresh kept`);

    new Notice(`Excalidraw: ${parts.join(', ') || 'nothing to do'} — ${file.basename}`);
  }

  // ── Settings ───────────────────────────────────────────────────────────────

  async loadSettings() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
  }

  async saveSettings() {
    await this.saveData(this.settings);
  }
}

class ExcalidrawDownloaderSettingTab extends PluginSettingTab {
  plugin: ExcalidrawDownloaderPlugin;

  constructor(app: App, plugin: ExcalidrawDownloaderPlugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display() {
    const { containerEl } = this;
    containerEl.empty();

    new Setting(containerEl)
      .setName('Attachment folder override')
      .setDesc(
        'Override the vault\'s "Default location for new attachments" setting. ' +
        'Use the same format as Obsidian: empty = vault root, /folder = fixed folder, ./ = same as note, ./folder = subfolder. ' +
        'Leave empty to read from vault config automatically.'
      )
      .addText(text =>
        text
          .setPlaceholder('(leave empty to use vault config)')
          .setValue(this.plugin.settings.attachmentFolderOverride)
          .onChange(async value => {
            this.plugin.settings.attachmentFolderOverride = value.trim();
            await this.plugin.saveSettings();
          })
      );
  }
}
