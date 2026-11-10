---
title: "Presets"
description: "Save PIrateRF transmission settings as named presets per mode, reload them with one click, and back them up as plain JSON files on the Pi."
nav_order: 5
---

# Presets

Stop typing the same parameters every time. A preset is a named snapshot of one mode's settings: frequency, audio file, RDS text, FT8 message and so on.

## Using presets in the web UI

1. Configure the transmission settings for a mode.
2. Click **📝 New** and give the preset a name.
3. Select the preset in the dropdown. That enables **💾 Save**, **🔃 Reload** and **✏️ Edit**.
4. Change settings and click **💾 Save** to update the preset.
5. Click **🔃 Reload** to throw away unsaved changes and go back to the preset's values.
6. Click **✏️ Edit** to rename or delete the preset.

Each mode has its own presets, and the UI remembers the last preset you selected across page reloads.

## Where presets live

Presets are JSON files under `files/presets/<module>/<name>.json` in the PIrateRF files directory (`PIRATERF_FILESDIR`, see [Configuration](configuration.md)). `<module>` is the gorpitx module name from the [Transmission modes](transmission-modes.md) table, for example `pifmrds` for the FM station. You can copy these files off the Pi to back them up or share them, and drop them back in to restore them.

The browser reads the preset list from the JSON directory index at `/files/presets/<module>/`. Loading, saving, renaming and deleting go over the WebSocket as the `preset.*` events in the [API](api.md#presets).
