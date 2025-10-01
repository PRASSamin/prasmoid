package consts

var UsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{- if .HasExample }}

Examples:
  {{.Example}}{{end -}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}{{- $printed := false}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
{{- if not $printed}}{{$printed = true}}

{{$group.Title}}:{{end}}  
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Available Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end -}}{{if .HasAvailableInheritedFlags }}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end -}}{{if .HasHelpSubCommands }}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var MAIN_QML = `import QtQuick 6.5
import QtQuick.Layouts 6.5
import org.kde.kirigami 2.20 as Kirigami
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.plasmoid 2.0

PlasmoidItem {
    id: root

    Plasmoid.backgroundHints: PlasmaCore.Types.NoBackground

    RowLayout {
        id: rowLayout
        anchors.centerIn: root

        Text {
            text: i18n("Hello from {{.Name}}!")
            color: "white"
            font.pointSize: 18
        }
    }
}`

var MAIN_XML = `<kcfg>
  <group name="General">
    <entry name="exampleOption" type="Bool">
      <default>true</default>
    </entry>
  </group>
</kcfg>
`

var GITIGNORE = `# Build artifacts
build/
*.plasmoid

# IDE files
.vscode/
.idea/

# OS-specific
.DS_Store
`

var PRASMOID_DTS = `/**
 * This file provides type definitions for the custom command environment in Prasmoid.
 * It is used by code editors like VS Code to provide autocompletion and type-checking.
 *
 * @see https://www.typescriptlang.org/docs/handbook/jsdoc-supported-types.html
 */

/**
 * The context object passed to every custom command's Run function.
 */
interface CommandContext {
  /**
   * Returns the command-line arguments passed after the command name.
   * @returns {string[]} An array of arguments.
   * @example
   * const args = ctx.Args();
   * console.log(args[0]);
   */
  Args(): string[];

  /**
   * Provides access to the flags passed to the command.
   */
  Flags(): {
    /**
     * Retrieves the value of a specific flag.
     * @param name The name of the flag to retrieve.
     * @returns {string | boolean | undefined} The value of the flag, or undefined if not found.
     * @example
     * const name = ctx.Flags().get("name");
     */
    get(name: string): string | boolean | undefined;

    /**
     * You can also access flags directly as properties, but using get() is recommended for better type safety.
     * @example
     * const myFlag = ctx.Flags().myFlagName; // Type is 'any'
     */
    [key: string]: any;
  };
}

/**
 * The global prasmoid module for interacting with the project.
 */
declare module "prasmoid" {
  /**
   * Retrieves a value from the project's metadata.json file.
   * @param key The key from the "KPlugin" section of metadata.json (e.g., "Id", "Version").
   * @returns {string | undefined} The value from the metadata, or undefined if not found.
   */
  export function getMetadata(key: string): string | undefined;
  /**
   * Registers a custom command.
   * @param config The configuration for the command.
   */
  export function Command(config: Config): void;
}

/**
 * Configuration for the custom command.
 */
interface Config {
  run: (ctx: CommandContext) => void;
  /** A brief description of your command. */
  short: string;
  /** A longer description that spans multiple lines. */
  long: string;
  /** Optional aliases for the command. */
  alias?: string[];
  /** Flag definitions for the command. */
  flags?: {
    name: string;
    shorthand?: string;
    type: "string" | "boolean";
    default?: string | boolean;
    description: string;
  }[];
}

interface Console {
  /**
   * Logs a red-colored message.
   */
  red(...message: any[]): void;

  /**
   * Logs a green-colored message.
   */
  green(...message: any[]): void;

  /**
   * Logs a yellow-colored message.
   */
  yellow(...message: any[]): void;

  /**
   * Flexible color logger. Last argument must be a color string.
   * @example
   * console.color("Hey", "you!", "red")
   */
  color(
    ...message: any[],
    color: "red" | "green" | "yellow" | "blue" | "magenta" | "cyan" | "black"
  ): void;
}

declare var console: Console;

type LocaleCode =
  | "af"
  | "ar"
  | "ast"
  | "az"
  | "be"
  | "bg"
  | "bn"
  | "br"
  | "bs"
  | "ca"
  | "ca@valencia"
  | "cs"
  | "cy"
  | "da"
  | "de"
  | "el"
  | "en"
  | "en_GB"
  | "eo"
  | "es"
  | "et"
  | "eu"
  | "fa"
  | "fi"
  | "fr"
  | "fy"
  | "ga"
  | "gd"
  | "gl"
  | "gu"
  | "he"
  | "hi"
  | "hr"
  | "hsb"
  | "hu"
  | "ia"
  | "id"
  | "is"
  | "it"
  | "ja"
  | "ka"
  | "kk"
  | "km"
  | "kn"
  | "ko"
  | "lt"
  | "lv"
  | "mai"
  | "mk"
  | "ml"
  | "mr"
  | "ms"
  | "nb"
  | "nds"
  | "ne"
  | "nl"
  | "nn"
  | "oc"
  | "pa"
  | "pl"
  | "pt"
  | "pt_BR"
  | "ro"
  | "ru"
  | "rw"
  | "se"
  | "si"
  | "sk"
  | "sl"
  | "sq"
  | "sr"
  | "sr@ijekavian"
  | "sr@ijekavianlatin"
  | "sr@latin"
  | "sv"
  | "ta"
  | "te"
  | "tg"
  | "th"
  | "tr"
  | "ug"
  | "uk"
  | "uz"
  | "uz@cyrillic"
  | "vi"
  | "wa"
  | "xh"
  | "zh_CN"
  | "zh_HK"
  | "zh_TW";

type PrasmoidConfig = {
  commands: {
    dir: string;
    ignore: string[];
  };
  i18n: {
    dir: string;
    locales: LocaleCode[];
  };
};
`

var PRASMOID_SVG = `<svg enable-background="new 0 0 128 128" viewBox="0 0 128 128" xmlns="http://www.w3.org/2000/svg"><linearGradient id="a" x1="64" x2="64" y1="4.3333" y2="124.43" gradientUnits="userSpaceOnUse"><stop stop-color="#80D8FF" offset="0"/><stop stop-color="#36C1FF" offset=".5888"/><stop stop-color="#00B0FF" offset=".9954"/></linearGradient><path d="m76.41 44.85 10.22-10.22c3.12-3.12 3.12-8.19 0-11.31l-16.97-16.98c-3.12-3.12-8.19-3.12-11.31 0l-16.98 16.97c-3.12 3.12-3.12 8.19 0 11.31l10.22 10.22c1.17 1.17 2.92 1.49 4.44 0.83 2.44-1.07 5.13-1.67 7.97-1.67s5.53 0.6 7.97 1.67c1.51 0.67 3.27 0.34 4.44-0.82z" fill="url(#a)"/><linearGradient id="d" x1="102.99" x2="102.99" y1="4.3333" y2="124.43" gradientUnits="userSpaceOnUse"><stop stop-color="#80D8FF" offset="0"/><stop stop-color="#36C1FF" offset=".5888"/><stop stop-color="#00B0FF" offset=".9954"/></linearGradient><path d="m121.66 58.34-16.97-16.97c-3.12-3.12-8.19-3.12-11.31 0l-10.23 10.22c-1.17 1.17-1.49 2.92-0.83 4.44 1.08 2.44 1.68 5.13 1.68 7.97s-0.6 5.53-1.67 7.97c-0.66 1.51-0.34 3.27 0.83 4.44l10.22 10.22c3.12 3.12 8.19 3.12 11.31 0l16.97-16.97c3.12-3.13 3.12-8.19 0-11.32z" fill="url(#d)"/><linearGradient id="c" x1="25.007" x2="25.007" y1="4.3333" y2="124.43" gradientUnits="userSpaceOnUse"><stop stop-color="#80D8FF" offset="0"/><stop stop-color="#36C1FF" offset=".5888"/><stop stop-color="#00B0FF" offset=".9954"/></linearGradient><path d="m44.85 51.59-10.22-10.22c-3.12-3.12-8.19-3.12-11.31 0l-16.98 16.97c-3.12 3.12-3.12 8.19 0 11.31l16.97 16.97c3.12 3.12 8.19 3.12 11.31 0l10.22-10.22c1.17-1.17 1.49-2.92 0.83-4.44-1.07-2.43-1.67-5.12-1.67-7.96s0.6-5.53 1.67-7.97c0.67-1.51 0.34-3.27-0.82-4.44z" fill="url(#c)"/><path d="m51.59 83.15-10.22 10.22c-3.12 3.12-3.12 8.19 0 11.31l16.97 16.97c3.12 3.12 8.19 3.12 11.31 0l16.97-16.97c3.12-3.12 3.12-8.19 0-11.31l-10.21-10.22c-1.17-1.17-2.92-1.49-4.44-0.83-2.44 1.08-5.13 1.68-7.97 1.68s-5.53-0.6-7.97-1.67c-1.51-0.67-3.27-0.34-4.44 0.82z" fill="url(#a)"/><linearGradient id="b" x1="64" x2="64" y1="48.833" y2="81.844" gradientUnits="userSpaceOnUse"><stop stop-color="#42A5F5" offset="0"/><stop stop-color="#1976D2" offset="1"/></linearGradient><circle cx="64" cy="64" r="16" fill="url(#b)"/><g opacity=".2"><path d="m64 7c1.34 0 2.59 0.52 3.54 1.46l16.97 16.97c0.94 0.94 1.46 2.2 1.46 3.54s-0.52 2.59-1.46 3.54l-10.22 10.22c-0.19 0.19-0.43 0.29-0.7 0.29-0.14 0-0.28-0.03-0.41-0.09-2.91-1.28-6-1.93-9.18-1.93s-6.27 0.65-9.18 1.93c-0.13 0.06-0.27 0.09-0.41 0.09-0.26 0-0.51-0.1-0.7-0.29l-10.22-10.22c-1.95-1.95-1.95-5.12 0-7.07l16.97-16.98c0.95-0.94 2.2-1.46 3.54-1.46m0-3c-2.05 0-4.09 0.78-5.66 2.34l-16.97 16.97c-3.12 3.12-3.12 8.19 0 11.31l10.22 10.22c0.76 0.76 1.78 1.17 2.82 1.17 0.55 0 1.1-0.11 1.62-0.34 2.44-1.07 5.13-1.67 7.97-1.67s5.53 0.6 7.97 1.67c0.52 0.23 1.07 0.34 1.62 0.34 1.04 0 2.06-0.4 2.82-1.17l10.22-10.22c3.12-3.12 3.12-8.19 0-11.31l-16.97-16.97c-1.57-1.56-3.61-2.34-5.66-2.34z" fill="#424242"/></g><g opacity=".2"><path d="m99.03 42.03c1.34 0 2.59 0.52 3.54 1.46l16.97 16.97c0.94 0.94 1.46 2.2 1.46 3.54s-0.52 2.59-1.46 3.54l-16.97 16.97c-0.94 0.94-2.2 1.46-3.54 1.46s-2.59-0.52-3.54-1.46l-10.22-10.22c-0.29-0.29-0.37-0.73-0.2-1.11 1.28-2.91 1.93-6 1.93-9.18s-0.65-6.27-1.93-9.18c-0.17-0.38-0.09-0.82 0.2-1.11l10.22-10.22c0.95-0.94 2.2-1.46 3.54-1.46m0-3c-2.05 0-4.09 0.78-5.66 2.34l-10.22 10.22c-1.17 1.17-1.49 2.92-0.83 4.44 1.08 2.44 1.68 5.13 1.68 7.97s-0.6 5.53-1.67 7.97c-0.66 1.51-0.34 3.27 0.83 4.44l10.22 10.22c1.56 1.56 3.61 2.34 5.66 2.34s4.09-0.78 5.66-2.34l16.97-16.97c3.12-3.12 3.12-8.19 0-11.31l-16.97-16.97c-1.58-1.57-3.62-2.35-5.67-2.35z" fill="#424242"/></g><g opacity=".2"><path d="m28.97 42.03c1.34 0 2.59 0.52 3.54 1.46l10.22 10.22c0.29 0.29 0.37 0.73 0.2 1.11-1.28 2.91-1.93 6-1.93 9.18s0.65 6.27 1.93 9.18c0.17 0.38 0.09 0.82-0.2 1.11l-10.22 10.22c-0.94 0.94-2.2 1.46-3.54 1.46s-2.59-0.52-3.54-1.46l-16.97-16.97c-0.94-0.95-1.46-2.2-1.46-3.54s0.52-2.59 1.46-3.54l16.97-16.97c0.95-0.94 2.21-1.46 3.54-1.46m0-3c-2.05 0-4.09 0.78-5.66 2.34l-16.97 16.97c-3.12 3.12-3.12 8.19 0 11.31l16.97 16.97c1.56 1.56 3.61 2.34 5.66 2.34s4.09-0.78 5.66-2.34l10.22-10.22c1.17-1.17 1.49-2.92 0.83-4.44-1.08-2.43-1.68-5.12-1.68-7.96s0.6-5.53 1.67-7.97c0.66-1.51 0.34-3.27-0.83-4.44l-10.21-10.22c-1.56-1.56-3.61-2.34-5.66-2.34z" fill="#424242"/></g><g opacity=".2"><path d="m73.59 84.99c0.26 0 0.51 0.1 0.7 0.29l10.22 10.22c0.94 0.94 1.46 2.2 1.46 3.54s-0.52 2.59-1.46 3.54l-16.97 16.97c-0.94 0.94-2.2 1.46-3.54 1.46s-2.59-0.52-3.54-1.46l-16.97-16.97c-0.94-0.94-1.46-2.2-1.46-3.54s0.52-2.59 1.46-3.54l10.22-10.22c0.19-0.19 0.43-0.29 0.7-0.29 0.14 0 0.28 0.03 0.41 0.09 2.91 1.28 6 1.93 9.18 1.93s6.27-0.65 9.18-1.93c0.13-0.06 0.27-0.09 0.41-0.09m0-3c-0.55 0-1.1 0.11-1.62 0.34-2.44 1.07-5.13 1.67-7.97 1.67s-5.53-0.6-7.97-1.67c-0.52-0.23-1.07-0.34-1.62-0.34-1.04 0-2.06 0.4-2.82 1.17l-10.22 10.21c-3.12 3.12-3.12 8.19 0 11.31l16.97 16.97c1.56 1.56 3.61 2.34 5.66 2.34s4.09-0.78 5.66-2.34l16.97-16.97c3.12-3.12 3.12-8.19 0-11.31l-10.22-10.22c-0.77-0.76-1.78-1.16-2.82-1.16z" fill="#424242"/></g><g opacity=".2"><path d="m64 51c7.17 0 13 5.83 13 13s-5.83 13-13 13-13-5.83-13-13 5.83-13 13-13m0-3c-8.84 0-16 7.16-16 16s7.16 16 16 16 16-7.16 16-16-7.16-16-16-16z" fill="#424242"/></g></svg>`

var JS_COMMAND_TEMPLATE = `/// <reference path="%s" />
const prasmoid = require("prasmoid");

prasmoid.Command({
    run: (ctx) => {
		const plasmoidId = prasmoid.getMetadata("Id");
		if (!plasmoidId) {
			console.red(
			"Could not get Plasmoid ID. Are you in a valid project directory?"
			);
			return;
		}

		console.color('%s Called', "blue");
	},
	short: "A brief description of your command.",
	long: "A longer description that spans multiple lines and likely contains examples\nand usage of using your command. For example:\n\nPlasmoid CLI is a CLI tool for KDE Plasmoid development.\nIt's a all-in-one tool for plasmoid development.",
	flags: [],
});`
