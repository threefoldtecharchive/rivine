// fillBlock populates the information fields in the block being
// presented.
function fillBlock(height) {
	var request = new XMLHttpRequest();
	var reqString = '/explorer/blocks/' + height;
	request.open('GET', reqString, false);
	request.send();
	var infoBody = document.getElementById('dynamic-elements');
	if (request.status != 200) {
		appendHeading(infoBody, 'Block Not Found');
		appendHeading(infoBody, 'Height: ' + height);
	} else {
		var explorerHash = JSON.parse(request.responseText);
		appendExplorerBlock(infoBody, explorerHash.block, explorerHash.unconfirmed!==true);
	}
}

// parseBlockQuery parses the query string in the URL and loads the block
// that makes sense based on the result.
function parseBlockQuery() {
	var urlParams;
	(window.onpopstate = function () {
	var match,
		pl     = /\+/g,  // Regex for replacing addition symbol with a space
		search = /([^&=]+)=?([^&]*)/g,
		decode = function (s) { return decodeURIComponent(s.replace(pl, ' ')); },
		query  = window.location.search.substring(1);
	urlParams = {};
	while (match = search.exec(query))
		urlParams[decode(match[1])] = decode(match[2]);
	})();
	fillBlock(urlParams.height);
}
parseBlockQuery();
