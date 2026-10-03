---
title: I Brought My Favorite Pi Extension to Claude Code
date: 2026-10-03
description: I missed Pi’s /answer workflow in Claude Code, so when mod support arrived on October 1, I built a version for it.
---

Claude Code added mod support on October 1, and after reading [Claude’s introduction to Claude Code mods](https://claude.dev/blog/getting-started-with-claude-code-mods), I knew what I wanted to try first: bringing over `/answer`, one of my favorite Pi extensions.

I subscribed to Claude about a week ago, after spending the past year trying different providers and open-source agent harnesses like OpenCode and Pi. Pi is the one I’ve used most. It gives me a small set of tools to start with, and I can extend it to work the way I like.

That’s what keeps me coming back to it. When something’s missing, I can ask an agent to build it. I’ve made a few extensions for my own workflow and shared them in [pi-extensions](https://github.com/PeteChu/pi-extensions).

One of my favorites is `/answer`. It pulls questions out of the agent’s last message and puts them in a UI where I can answer each one separately. I don’t have to squeeze all my replies into the prompt box or worry about accidentally sending half an answer while trying to start a new line.

It didn’t take long to miss that in Claude Code. There’s already a built-in `AskUserQuestion` dialog, but Claude Code doesn’t always invoke it when questions appear in a regular message. Sometimes that leaves me typing all my answers into the prompt.

I asked Claude to build a version of `/answer` for Claude Code.

The mod sends Claude Code’s latest response to Haiku, which extracts the questions as JSON. It then passes them to the existing `AskUserQuestion` UI. No separate interface to learn—just a way to open that dialog when I need it.

<!-- Add a screenshot of the mod's question dialog here. -->

You can install [the mod](https://github.com/petechu/cc-answer) with:

```bash
/plugin marketplace add PeteChu/cc-answer
/plugin install answer@cc-answer
/reload-plugins
```

Then run `/answer` whenever you’d like to answer the questions in Claude Code’s latest response through the dialog.

Mod support is only two days old as I write this, and I’ve already got one of my favorite Pi workflows back. How would you mod your Claude Code setup?
