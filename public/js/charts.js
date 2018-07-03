var ticketsFunc =function(gData){
    ticketsPrice = [];
    gData.time.forEach((n, i) => {
        ticketsPrice.push([new Date (n*1000), gData.valuef[i]]);
    });
    return ticketsPrice
}

var difficultyFunc = function(gData){
    difficulty = [];
    gData.time.forEach((n, i) => {
        difficulty.push([new Date(n*1000), gData.sizef[i]]);
    });
    return difficulty
}

var supplyFunc = function (gData){
    coinSupply = [];
    gData.time.forEach((n, i) => {
        coinSupply.push([new Date(n*1000), gData.valuef[i]])
    });
    return coinSupply
}

var timeBtwBlocksFunc = function(gData){
    timeBtwBlocks = [];
    gData.value.forEach((n, i) => {
        if (n === 0) {return}
        timeBtwBlocks.push([n, gData.valuef[i]]);
    });
    return timeBtwBlocks
}

var blockSizeFunc = function(gData){
    blockSize = [];
    gData.time.forEach((n, i) => {
        blockSize.push([new Date(n*1000), gData.size[i]]);
    });
    return blockSize
}

var blockChainSizeFunc = function(gData){
    blockchainSize = [];
    gData.time.forEach((n, i) => {
        blockchainSize.push([new Date(n*1000), gData.chainsize[i]]);
    });
    return blockchainSize
}

var txPerBlockFunc = function(gData){
    txPerBlock = [];
    gData.value.forEach((n, i) => {
        txPerBlock.push([n, gData.count[i]]);
    });
    return txPerBlock
}

var txPerDayFunc = function (gData){
    txPerDay = [];
    gData.timestr.forEach((n, i) => {
        txPerDay.push([new Date(n), gData.count[i]]);
    });
    return txPerDay
}

var poolSizeFunc = function(gData){
    poolSize = [];
    gData.time.forEach((n, i) =>{
        poolSize.push([new Date(n*1000), gData.sizef[i]]);
    });
    return poolSize
}

var poolValueFunc = function(gData) {
    poolValue = [];
    gData.time.forEach((n, i) => {
        poolValue.push([new Date(n*1000), gData.valuef[i]]);
    })
    return poolValue
}

var blockFeeFunc = function (gData){
    feePerBlock = [];
    gData.count.forEach((n,i) => {
        feePerBlock.push([n, gData.sizef[i]]);
    }); 
    return feePerBlock
}

var ticketSpendTypeFunc = function(gData) {
    ticketSpendType = [];
    gData.height.forEach((n,i) => {
        ticketSpendType.push([n, gData.unspent[i], gData.revoked[i], gData.voted[i]]);
    }); 
    return ticketSpendType
}

var ticketByOutputCountFunc = function(gData) {
    ticketByOutputCount = [];
    gData.height.forEach((n,i) => {
        ticketByOutputCount.push([n, gData.solo[i], gData.pooled[i], gData.txsplit[i]]);
    }); 
    return ticketByOutputCount
}

var mapDygraphOptions = function (data,labelsVal, isDrawPoint, yLabel, xLabel, titleName, labelsMG, labelsMG2){
    return { 
            'file': data,
            digitsAfterDecimal: 8,
            labels: labelsVal,
            drawPoints: isDrawPoint,
            ylabel: yLabel,
            xlabel: xLabel,
            labelsKMB: labelsMG,
            labelsKMG2: labelsMG2,
            title: titleName,
            fillGraph: false,
            stackedGraph: false,
            legendFormatter: Formatter,
            plotter: Dygraph.Plotters.linePlotter,
            colors: ['rgb(0,128,127)']
        }
}

var getAPIData = function (gType, callback){
    $.ajax({
        type: "GET",
        url: '/api/chart/'+gType,
        beforeSend: function() {},
        success: function(data) {
            callback(gType, data);
        }
    });
}

function Formatter(data) {
    if (data.x == null) return '';
    var html = this.getLabels()[0] + ': ' + data.xHTML;
    data.series.forEach(function(series){
        var labeledData = ' <span style="color: ' + series.color + ';">' +series.labelHTML + ': ' + series.yHTML;
        html += '<br>' + series.dashHTML  + labeledData +'</span>';
    });
    return html;
}

 function darkenColor(colorStr) {
   var color = Dygraph.toRGB_(colorStr);
   color.r = Math.floor((255 + color.r) / 2);
   color.g = Math.floor((255 + color.g) / 2);
   color.b = Math.floor((255 + color.b) / 2);
   return 'rgb(' + color.r + ',' + color.g + ',' + color.b + ')';
}

function barchartPlotter(e) {
   var ctx = e.drawingContext;
   var points = e.points;
   var y_bottom = e.dygraph.toDomYCoord(0);

   ctx.fillStyle = darkenColor(e.color);

   var min_sep = Infinity;
   for (var i = 1; i < points.length; i++) {
   var sep = points[i].canvasx - points[i - 1].canvasx;
   if (sep < min_sep) min_sep = sep;
   }
   var bar_width = Math.floor(2.0 / 3 * min_sep);

   for (var i = 0; i < points.length; i++) {
   var p = points[i];
   var center_x = p.canvasx;

   ctx.fillRect(center_x - bar_width / 2, p.canvasy,
       bar_width, y_bottom - p.canvasy);

   ctx.strokeRect(center_x - bar_width / 2, p.canvasy,
       bar_width, y_bottom - p.canvasy);
   }
}
