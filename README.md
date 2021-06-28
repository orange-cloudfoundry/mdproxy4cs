# MetaData Proxy 4 CloudStack

`mdproxy4cs` is installed in our BOSH stemcell where it is used
to abstract the complexity to access the IaaS metadata from CloudStack
needed by the bosh-agent to finalise the virtual machines configuration.

Workflow:

- bosh-agent first boot: setup a static configuration inside the VM
- `mdproxy4cs` start!

  - set up a predetermined Link-local IP address and listen on the defined port
    (`MDPROXY4CS_HTTP_LISTEN`)
  - discover the IaaS IP address serving the metadata
    (`MDPROXY4CS_INAME`)

- bosh-agent second boot: request the VM metadata to the IaaS
  through `mdproxy4cs`

## Usage

Installation example:

```sh
tempdir=$(mktemp -d)
repo="orange-cloudfoundry/mdproxy4cs"
version=$(curl -fsL "https://api.github.com/repos/${repo}/releases/latest" | sed -n -e 's/.*"tag_name":[[:space:]]*"v\([^"]*\)".*/\1/p')
name="mdproxy4cs_${version}_linux_amd64"

curl -fsL "https://github.com/${repo}/releases/download/v${version}/${name}.tar.gz" |
tar -xz -C "$tempdir"

install -m 0755 -D "${tempdir}/${name}/mdproxy4cs" /usr/bin/mdproxy4cs
install -m 0644 -D "${tempdir}/${name}/assets/default" /etc/default/mdproxy4cs
install -m 0644 -D "${tempdir}/${name}/assets/mdproxy4cs.service" /usr/share/mdproxy4cs/mdproxy4cs.service
install -m 0755 -D "${tempdir}/${name}/assets/pre-start.sh" /usr/share/mdproxy4cs/pre-start.sh

systemctl enable /usr/share/mdproxy4cs/mdproxy4cs.service

rm -rf "${tempdir}"
```

## Why?

The metadata are provided by the CloudStacks virtual routers.
Before CloudStack 4.14 (more exactly the patch:
[#3587](https://github.com/apache/cloudstack/pull/3587)
vRouter in redundant mode acquire guest IP from first IP of the tier),
the IP addresses of the vRouters was not set.
They could be found through DHCP requests,
but changed over time
(when a vRouter was recreated and the bosh-agent restarted for one case,
or the connectivity to the BOSH director is temporarily lost).

This made the process of acquiring the metadata complex and unreliable;
thus this middleware was made.

As of today, we are investigating to replace this component by a more
versatile and widely adopted one,
such as [cloud-init](https://github.com/canonical/cloud-init/).

Stay tuned for more.

## Related resources

- <https://www.cloudfoundry.org/>,
  <https://github.com/cloudfoundry/bosh-linux-stemcell-builder>
- <https://cloudstack.apache.org/>
  (<https://github.com/apache/cloudstack/>)

  - <http://docs.cloudstack.apache.org/en/4.14.0.0/adminguide/api.html#user-data-and-meta-data>
  - <http://docs.cloudstack.apache.org/en/4.14.0.0/adminguide/virtual_machines/user-data.html#user-data-and-meta-data>
