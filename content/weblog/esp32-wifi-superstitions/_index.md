---
draft: false
date: 2025-03-15T12:00:00+00:00
title: ESP32 WiFi superstitions
tags: esp32, wifi, networking, superstition
aliases:
  - esp32-wifi-superstitions.html
---

The [ESP32](https://www.espressif.com/en/products/socs/esp32) is a popular microcontroller to use for do-it-yourself home automation, sensors, and a variety of other bits and bobs that you might want to take care of around the house. It's the successor to the venerable [ESP8266](https://www.espressif.com/en/products/socs/esp8266) which has found its way into many of our WiFi connected devices (seriously, open up a device and chances are relatively large that you'll find one).

After having done quite a few projects based on the ESP32 in the [Arduino](https://www.arduino.cc/) and [esp-idf](https://docs.espressif.com/projects/esp-idf/en/stable/esp32/get-started/index.html) frameworks I did start to notice some pecularities with my deployed devices (fancy wording for the one in my electrical cabinet that's *supposed* to send the electrical usage data somewhere and a hodgepodge of sensor boards around the house, mostly the *excellent* [Snuffelaar](https://revspace.nl/Snuffelaar) by [Sebastius](https://tweakers.net/reviews/8876/tweaker-sebastius-over-zijn-soldeerworkshops-en-reparatieprojecten.html) with firmware written by [Juerd](https://github.com/juerd)).

It seemed some of my ESP32 based boards were regularly losing connectivity. Initially my thoughts went out to the terrible power supplies I was using to run them (the cheapest of the cheap USB power supplies that came with a variety of accessories around my house). After switching these out for some more accessible and probably better quality tested [Ikea chargers](https://www.ikea.com/nl/en/p/smahagel-3-port-usb-charger-white-60539177/) the problems, however, persisted.

Asking around for experience from others at [RevSpace](https://revspace.nl/), my local hackerspace, seemed to indicate that people had seen similar things with their ESP32-based projects. But not everyone had these issues. Slowly I started gathering more and more "superstitions" around how to keep these microcontrollers connected to my internet. Here are my favorite ones, I have applied all of these and while I haven't tested them one-by-one the combination of them has ensured steady connections on my SSIDs.

While these workarounds don't quite come close to placing a hexagon of CR2023 batteries around your ESP32 while you chant the 802.11ax specification at it, they have no basis in any *actual* research I did. Take these as anecdotal workarounds for ESP32's losing connectivity to your WiFi.

## Turn off power saving on the ESP32

The ESP8266 never had any power saving for its WiFi modem stack, however the ESP32 *does*. To me this is the most likely culprit in that in some network configurations, perhaps in combination with some radios, the power saving does something that makes it stop interacting with the network.

In your personal handcrafted firmware you can [use the following](https://docs.espressif.com/projects/esp-idf/en/stable/esp32/api-reference/network/esp_wifi.html#_CPPv415esp_wifi_set_ps14wifi_ps_type_t), which should work in esp-idf *and* Arduino (from what I've been told):

```c
esp_wifi_set_ps(WIFI_PS_NONE)
```

For [ESPHome](https://esphome.io/) based projects you can add:

```yaml
wifi:
    power_save_mode: NONE
```

## Set your APs to use 20 Mhz wide channels

If you have fancy network hardware then you can likely configure the channel width for the network that serves your ESP32's. From what people and the internet tell me you *should* set the band width on the 2.4 Ghz network that your boards use to **20 Mhz**, not 40, not 60, and definitely not automatic.

## Pin your ESP32's to a single AP

It seems that when an ESP32 connects it goes straight for the first access point it sees. No matter if that access point is not the one you've taped it to. This can lead to bad connectivity, especially since I've not really observed ESP32's moving around to other access points. If your network hardware allows it, you should pin the device to the closest one.

---

These urban legends have so far made it seem that at least my problems are ghosts of the past. I haven't had a device drop from the network in about a week or two now while they used to drop multiple times per day. I'm planning to drop my application level keep-alives (still a good idea, I'll write about them another time) because they seem to not be necessary at all anymore.

I hope these are of help to anyone, and that the spirits of the ESP32 deem your network worthy too.
