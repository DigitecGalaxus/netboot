# IPXE Menu Generator

This container consumes the `NETBOOT_SERVER_IP` that is passed via an environment variable and produces an IPXE menu. It takes the templates and generates the menu based on the files available in the `assets` folder. It checks for a `dev` and `prod` folder and generates the menu based on the files in there. The menu is then written to the `config/menus` folder.

Based on the available variables passed in [ipxe-menu-generator.env](./ipxe-menu-generator.env), the following workflow will be impacted.

## IPXE Workflow

The IPXE workflow is as follows:

1. The next-server defined on the DHCP server points to either a local netboot server per location or the azure netboot server
2. The netboot server serves the IPXE binary (`undionly.kpxe`, `ipxe32.efi`, `ipxe64.efi`), which points to the `menu.ipxe` which is dyanmically generated on each netboot server.
3. If the `next-server` equals the value of the azure netboot server, it checks the availability of faster netboot server which is exposed through a reverse proxy. If the reverse proxy is available, the HTTP basic auth details will be created and the correct variables will be set (`http-protocol`, `url`, `basicAuth`).
4. If the check on the faster netboot server fails, a check will be made, if a storage account in azure, which holds the necessary files, is available. If it is available, the correct variables will be set (`http-protocol`, `url`, `sas_token`).
5. If none of the above is true, we will use the local netboot server as a primary source.
6. Based on the available variables, the menu will be generated and served to the client.

Note: `sas_token` and `basicAuth` are mutually exclusive. If one variable is not set, the IPXE menu generator will not parse the unused variable.

## MAC specific booting

The MAC specific booting can be done in IPXE using `chain --autofree tftp://${next-server}/MAC-${mac:hexraw}.ipxe`. Currently we do not make use of this feature. Make sure to also check the file permissions of the .ipxe file.

## Language support

The language support is done by evaluating the gateway address and statically setting the language. Once a gateway matches the location of a foreign language, we pass the `language` parameter to the kernel. The kernel then sets the language accordingly. The default language is `de_CH`. We use the language support for the operating system and the keyboard layout.

Those settings are defined in the [menu.ipxe](./menu.ipxe.j2) file:

```ipxe
# Lausanne
iseq ${netX/gateway} 172.20.72.1 && set language fr_CH && goto initial_menu

...

:startboot
imgfree
kernel ${kernel_url}vmlinuz ip=dhcp boot=casper netboot=url url=${squash_url} initrd=initrd locale=${language} ${cmdline}
initrd ${kernel_url}initrd
boot
```
