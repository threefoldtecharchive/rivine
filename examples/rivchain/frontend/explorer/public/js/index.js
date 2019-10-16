function fillGeneralStats() {
	var request = new XMLHttpRequest();
	request.open('GET', '/explorer', true);
	request.onload = function() {
		if (request.status !== 200) {
			return;
		}
		var explorerStatus = JSON.parse(request.responseText);

		var height = document.getElementById('height');
		linkHeight(height, explorerStatus.height);

		var blockID = document.getElementById('blockID');
		linkHash(blockID, explorerStatus.blockid);

		document.getElementById('difficulty').innerHTML = readableDifficulty(explorerStatus.difficulty);
	// 	document.getElementById('maturityTimestamp').innerHTML = formatUnixTime(explorerStatus.maturitytimestamp);
	// 	document.getElementById('totalCoins').innerHTML = readableCoins(explorerStatus.totalcoins);
 	};
	request.send();
}
fillGeneralStats();
