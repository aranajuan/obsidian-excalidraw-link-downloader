import 'obsidian';

declare module 'obsidian' {
  interface App {
    /**
     * Retrieve a vault configuration value.
     * @public
     * @since 1.4.0
     */
    getConfig(key: string): string | null;
  }
}
