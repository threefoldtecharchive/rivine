// Start by loading the required chart packages
google.charts.load('current', {'packages': ['line']});
google.charts.setOnLoadCallback(init);

function init() {
	showCurrentStats();
	// Load the graphs of the initial history value
	if (document.getElementById('history').value) {
		loadHistory();
	}
}

function showCurrentStats() {
	getLatestBlockFacts().then(function(facts) {
		setCurrentValues(facts);

		// Refresh every 2 mins
		setTimeout(showCurrentStats, 120 * 1000);
	});
}

function setCurrentValues(values) {
	document.getElementById('current-difficulty').innerHTML = values.difficulty+ ' BS';
	document.getElementById('current-height').innerHTML = values.height;
	document.getElementById('current-bs').innerHTML = values.estimatedactivebs + ' BS';
}

function loadRange() {
	var start = document.getElementById('range-start').value;
	var end = document.getElementById('range-end').value;

	getRangeStats(start, end).then(function(stats) {
		drawCharts(stats);
	});
}

function loadHistory() {
	var history = document.getElementById('history').value;

	getHistoryStats(history).then(function(stats) {
		drawCharts(stats);
	})
}

function drawCharts(stats) {
	// Calculate the amount of data points we need to have a moving window average for roughly 10% of the supplied data points
	var avgBlocks = Math.round(stats.blockcount / 10);
	var timeHeight = [['Timestamp', 'Block height']];
	var blockTime = [['Block Height', 'Block creation time', 'Average block creation time of last ' + avgBlocks + ' blocks']];
	var activeBS = [['Timestamp', 'Active BS']];
	var blockCreatorDistribution = [['Address', 'Blocks Created']];
	var txnCount = [['Block Height', 'Transaction count']];
	var blockDifficulty = [['Block Height', 'Difficulty']];

	for (var i = 0; i < stats.blocktimestamps.length; i++) {
		stats.blocktimestamps[i] = new Date(stats.blocktimestamps[i] * 1000);
	}

	// Collect the block time and average block time
	var avgBlockTime = 0;
	for (var i = 0; i < stats.blocktimes.length; i++) {
		// Keep track of the average. For every data point, apply its increase to said average.
		// If we processesed sufficient data points to reach the actual value of the average (i.e.
		// after `avgBlocks` datapoints), substract the first value in what would be the series from
		// which the average is calculated. Now we could keep track of the sum of the original datapoints
		// and only divide to get the average when we actually push it in the table, which means we would
		// do only one division per loop iteration instead of 2, but when it comes to javascript I have
		// serious trust issues, so I prefer to keep the value as low as possible. If this additional
		// division causes performance issues on your machine, it's time to upgrade your hardware.
		avgBlockTime += stats.blocktimes[i] / avgBlocks;
		if (i < avgBlocks) {
			blockTime.push([stats.blockheights[i], stats.blocktimes[i], null]);
			continue
		}
		avgBlockTime -= stats.blocktimes[i - avgBlocks] / avgBlocks;
		blockTime.push([stats.blockheights[i], stats.blocktimes[i], avgBlockTime]);
	}

	// Collect linear stats
	for (var i = 0; i < stats.blockcount; i++) {
		timeHeight.push([stats.blocktimestamps[i], stats.blockheights[i]]);
		activeBS.push([stats.blocktimestamps[i], parseInt(stats.estimatedactivebs[i])]);
		txnCount.push([stats.blockheights[i], stats.blocktransactioncounts[i]]);
		blockDifficulty.push([stats.blockheights[i], parseInt(stats.difficulties[i])]);
	}

	// Collect the data for the block creator distribution pie
	Object.keys(stats.creators).forEach(function(key) {
		blockCreatorDistribution.push([key, stats.creators[key]]);
	});

	// Make sure graph container is displayed
	document.getElementById('graph-container').style.display = 'Block';

	// ***************
	// Render graphs
	// ***************
	var heightWrapper = new google.visualization.ChartWrapper({
		chartType: 'LineChart',
		dataTable: timeHeight,
		options: {explorer: {actions: ['dragToZoom', 'rightClickToReset'], keepInBounds: true, maxZoomIn: 0.01}, 'title': 'Chain Height', legend: {position: 'none'}, animation: {duration: 1000, easing: 'out', startup: true}},
		containerId: 'height-graph'
	});
	heightWrapper.draw();

	var creationTimeWrapper = new google.visualization.ChartWrapper({
		chartType: 'LineChart',
		dataTable: blockTime,
		options: {explorer: {actions: ['dragToZoom', 'rightClickToReset'], keepInBounds: true, maxZoomIn: 0.01}, 'title': 'Block Creation Time (seconds since previous block)', legend: {position: 'none'}, animation: {duration: 1000, easing: 'out', startup: true}},
		containerId: 'creationTime-graph',
	});
	creationTimeWrapper.draw();

	var activebsWrapper = new google.visualization.ChartWrapper({
		chartType: 'LineChart',
		dataTable: activeBS,
		options: {explorer: {actions: ['dragToZoom', 'rightClickToReset'], keepInBounds: true, maxZoomIn: 0.01}, 'title': 'Estimate Active Blockstakes', legend: {position: 'none'}, animation: {duration: 1000, easing: 'out', startup: true}},
		containerId: 'bs-graph'
	});
	activebsWrapper.draw();

	var bcdWrapper = new google.visualization.ChartWrapper({
		chartType: 'PieChart',
		dataTable: blockCreatorDistribution,
		options: {'title': 'Block Creator Distribution', legend: {position: 'right'}, animation: {duration: 1000, easing: 'out', startup: true}},
		containerId: 'bcd-graph'
	});
	bcdWrapper.draw();

	var txnCountWrapper = new google.visualization.ChartWrapper({
		chartType: 'LineChart',
		dataTable: txnCount,
		options: {explorer: {actions: ['dragToZoom', 'rightClickToReset'], keepInBounds: true, maxZoomIn: 0.01}, 'title': 'Block Transaction Count', legend: {position: 'none'}, animation: {duration: 1000, easing: 'out', startup: true}},
		containerId: 'txnCount-graph'
	});
	txnCountWrapper.draw();

	var difficultyWrapper = new google.visualization.ChartWrapper({
		chartType: 'LineChart',
		dataTable: blockDifficulty,
		options: {explorer: {actions: ['dragToZoom', 'rightClickToReset'], keepInBounds: true, maxZoomIn: 0.01}, 'title': 'Block Difficulty', legend: {position: 'none'}, animation: {duration: 1000, easing: 'out', startup: true}},
		containerId: 'difficulty-graph'
	});
	difficultyWrapper.draw();

	// ***************
	// Event Handlers
	// ***************
	google.visualization.events.addListener(heightWrapper, 'select', (e) => {
		var selection = heightWrapper.getChart().getSelection()[0];
		// Index 0 are the labels
		var row = selection.row + 1;
		// Block height is in the column at index 1
		var block = timeHeight[row][1];

		blockDetailPage(block);
	});

	google.visualization.events.addListener(creationTimeWrapper, 'select', (e) => {
		var selection = creationTimeWrapper.getChart().getSelection()[0];
		// Index 0 are the labels
		var row = selection.row + 1;
		// block heights are in the column at index 1
		var block = blockTime[row][0];

		blockDetailPage(block);
	});

	google.visualization.events.addListener(activebsWrapper, 'select', (e) => {
		var selection = activebsWrapper.getChart().getSelection()[0];
		// Since there is no block to link in this datatable,
		// load it from another datatable. This should not be a problem
		// as all datatables for line charts should have the same info if the
		// primary index matches.
		//
		// Index 0 are the labels
		var row = selection.row + 1;
		// Sanity check to see if the timestamp matches
		if (timeHeight[row][0] == activeBS[row][0]) {
			var block = blockTime[row][0];
			blockDetailPage(block);
		}
	});

	google.visualization.events.addListener(bcdWrapper, 'select', (e) => {
		var selection = bcdWrapper.getChart().getSelection()[0];

		var row = selection.row + 1;
		var address = blockCreatorDistribution[row][0];

		hashDetailPage(address);
	});

	google.visualization.events.addListener(txnCountWrapper, 'select', (e) => {
		var selection = txnCountWrapper.getChart().getSelection()[0];

		var row = selection.row + 1;
		var block = txnCount[row][0];

		blockDetailPage(block);
	});

	google.visualization.events.addListener(difficultyWrapper, 'select', (e) => {
		var selection = difficultyWrapper.getChart().getSelection()[0];

		var row = selection.row + 1;
		var block = blockDifficulty[row][0];

		blockDetailPage(block);
	})

	// scroll to the graphs
	document.getElementById('graph-container').scrollIntoView({'behavior': 'smooth', 'block': 'start'});
}

function blockDetailPage(block) {
	window.location.href = '/block.html?height=' + block;
}

function hashDetailPage(hash) {
	window.location.href = 'hash.html?hash=' + hash;
}

function getRangeStats(start, end) {
	return new Promise(function(resolve, reject) {
		var request = new XMLHttpRequest();
		request.open('GET', '/explorer/stats/range?start=' + start + '&end=' + end, true);
		request.onload = function() {
			if (request.status != 200) {
				console.log(request);
				reject(request.status);
			}
			resolve(JSON.parse(request.responseText));
		};
		request.send();
	})
}

function getHistoryStats(history) {
	return new Promise(function(resolve, reject) {
		var request = new XMLHttpRequest();
		request.open('GET', '/explorer/stats/history?history=' + history, true);
		request.onload = function() {
			if (request.status != 200) {
				reject(request.status);
			}
			resolve(JSON.parse(request.responseText));
		};
		request.send()
	})
}

function getLatestBlockFacts() {
	return new Promise(function(resolve, reject) {
		var request = new XMLHttpRequest();
		request.open('GET', '/explorer', true);
		request.onload = function() {
			if (request.status != 200) {
				reject(request.status);
			}
			resolve(JSON.parse(request.responseText));
		};
		request.send();
	})
}