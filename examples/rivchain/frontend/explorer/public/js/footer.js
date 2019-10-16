// addOptionalFooter adds an optional footer with the version content
function addOptionalFooter() {
	var version=null;
	var versionpath=null;
	if (version == null) {
		return;
	}
	var footer = document.getElementById('footer');

	var footerContentDiv = document.createElement('div');
	footerContentDiv.id = 'footer-content';
	footer.appendChild(footerContentDiv)

	var versionDiv = document.createElement('div');
	versionDiv.id = 'version';
	footerContentDiv.appendChild(versionDiv)

	var versionParagraph = document.createElement('p');
	versionParagraph.textContent = "Version: ";
	versionDiv.appendChild(versionParagraph);

	var commitLink = document.createElement('a')
	commitLink.textContent = version
	commitLink.href = "https://github.com/threefoldtech/rivine/examples/rivchain" + versionpath;
	versionParagraph.appendChild(commitLink);
}
addOptionalFooter();
