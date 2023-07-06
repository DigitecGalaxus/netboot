# IPXE Menu Generator

This container consumes the `NETBOOT_SERVER_IP` that is passed via an environment variable and produces an IPXE menu. It takes the templates and generates the menu based on the files available in the `assets` folder. It checks for a `dev` and `prod` folder and generates the menu based on the files in there. The menu is then written to the `config/menus` folder.

## MAC specific booting

The MAC specific booting can be done in IPXE using `chain --autofree tftp://${next-server}/MAC-${mac:hexraw}.ipxe`. Currently we do not make use of this feature.

## Language support

The language support is done by evaluating the gateway address and statically setting the language. Once a gateway matches the location of a foreign language, we pass the `language` parameter to the kernel. The kernel then sets the language accordingly. The default language is `de_CH`. We use the language support for the operating system and the keyboard layout.
