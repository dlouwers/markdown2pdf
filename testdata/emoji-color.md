# Color Emoji Test

This document tests color emoji rendering using Twemoji PNGs.

## Common Emoji (Should Render as Color PNGs)

### Smileys & People
😀 😃 😄 😁 😆 😅 🤣 😂 😊 😇 🙂 🙃 😉 😌 😍 🥰 😘 😗 😙 😚 😋 😛 😝 😜 🤪 🤨 🧐 🤓 😎

### Hand Gestures
👍 👎 👊 ✊ 🤛 🤜 🤞 ✌️ 🤟 🤘 👌 🤌 🤏 👈 👉 👆 👇 ☝️ ✋ 🤚 🖐 🖖 👋 🤙 💪 🦾 🖕 ✍️ 🙏

### Hearts & Symbols
❤️ 🧡 💛 💚 💙 💜 🖤 🤍 🤎 💔 ❣️ 💕 💞 💓 💗 💖 💘 💝 💟 ☮️ ✝️ ☪️ 🕉 ☸️ ✡️ 🔯 🕎 ☯️ ☦️

### Status & Symbols
✅ ❌ ⭐ 🌟 💯 🔥 💥 ✨ 💫 🎉 🎊 🎈 🎁 🎀 🏆 🥇 🥈 🥉 ⚡ 💧 🌈 ☀️ ⛅ ☁️ 🌙 ⭐

### Activities & Objects
🎮 🎯 🎲 🎰 🎪 🎨 🎭 🎬 🎤 🎧 🎵 🎶 🎹 🎺 🎸 🪕 🎻 🎷 🥁 📱 💻 ⌨️ 🖱 🖲 🕹 💾 💿

### Transport & Places
🚀 🛸 🚁 🛩 ✈️ 🛫 🛬 🪂 💺 🚂 🚃 🚄 🚅 🚆 🚇 🚈 🚉 🚊 🚝 🚞 🚋 🚌 🚍 🚎 🚐 🚑

## Mixed Text and Emoji

Regular text can contain emoji inline like this: Hello 👋 World 🌍! This is great 🎉 and works well ✨.

You can also have **bold text with emoji 💪** and _italic text with emoji 🎨_ and even `code with emoji 🚀`.

### In Lists

- Item with emoji 🎯
- Another item 🎮
- Third item 🎨
  - Nested with emoji 🎭
  - Another nested 🎪

1. Numbered item 🥇
2. Second place 🥈
3. Third place 🥉

### In Tables

| Emoji | Name | Category |
|-------|------|----------|
| 😀 | Grinning | Smiley |
| 👍 | Thumbs Up | Hand |
| ❤️ | Red Heart | Heart |
| 🚀 | Rocket | Transport |
| 🔥 | Fire | Symbol |

## Emoji in Links

Click here for [emoji link 🔗](https://example.com) or visit [rocket site 🚀](https://example.org).

## Edge Cases

### Multiple Emoji Together
🎉🎊🎈 No spaces between emoji

### Emoji at Line Boundaries
This is a very long line with emoji at the end to test wrapping behavior when we reach the right margin 🚀

### Mixed Scripts
English 🌍 日本語 🗾 Emoji 😊 中文 🇨🇳 العربية 🌙

## Non-Common Emoji (Should Fallback to Font)

These emoji are not in the common 100 list and should render using the Noto Emoji font:

🦄 🦖 🦕 🦐 🦑 🦞 🦀 🐡 🐠 🐟 🐬 🐳 🐋 🦈
