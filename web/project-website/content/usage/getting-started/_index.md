---
title: Getting Started
weight: 0
---

*This will be finalised when a release of Flamenco 3 can actually be downloaded.*

In broad terms, to render with Flamenco, follow these steps:

1. Download Flamenco (link will become available when we release the first beta version).
2. Create a directory on some storage, like a NAS, and make sure it's available at the same path on each computer.
3. Install Blender on each computer you want to render on. It should be in the same place everywhere.
4. Pick the computer that will manage the farm. Run `flamenco-manager` on it. This will start a web browser with the *Flamenco Setup Assistant*.
5. Step through the assistant, pointing it to the storage (step 3) and Blender (step 3). Be sure to confirm at the final step.
6. Download the *Blender add-on* and install it. The link is in the top-right corner in your browser.
7. Configure the add-on by giving it the address of Flamenco Manager. You can see this in your web browser, and the Flamenco Manager logs also show URLs you can try. Be sure to click the checkmark to check the connection.
8. Save your Blend file in the shared storage.
9. Tell Flamenco to render it. You can find the Flamenco panel in Blender's Output Properties.

Curious about [what changed since the last major release][what-is-new]?

[what-is-new]: {{< ref "what-is-new" >}}
