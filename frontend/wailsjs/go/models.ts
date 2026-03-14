export namespace backtest {
	
	export class Summary {
	    totalSignals: number;
	    executedTrades: number;
	    wins: number;
	    losses: number;
	    skippedSignals: number;
	    winRate: number;
	    finalBalance: number;
	    totalPnL: number;
	    roi: number;
	    maxDrawdown: number;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSignals = source["totalSignals"];
	        this.executedTrades = source["executedTrades"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.skippedSignals = source["skippedSignals"];
	        this.winRate = source["winRate"];
	        this.finalBalance = source["finalBalance"];
	        this.totalPnL = source["totalPnL"];
	        this.roi = source["roi"];
	        this.maxDrawdown = source["maxDrawdown"];
	    }
	}
	export class Trade {
	    id: number;
	    strategyName: string;
	    direction: string;
	    signalIndex: number;
	    entryIndex: number;
	    exitIndex: number;
	    signalTime: number;
	    entryTime: number;
	    exitTime: number;
	    entryPrice: number;
	    exitPrice: number;
	    stopLoss: number;
	    takeProfit: number;
	    quantity: number;
	    orderValueUSDT: number;
	    pnl: number;
	    pnlPercent: number;
	    balanceAfter: number;
	    exitReason: string;
	    holdBars: number;
	
	    static createFrom(source: any = {}) {
	        return new Trade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.strategyName = source["strategyName"];
	        this.direction = source["direction"];
	        this.signalIndex = source["signalIndex"];
	        this.entryIndex = source["entryIndex"];
	        this.exitIndex = source["exitIndex"];
	        this.signalTime = source["signalTime"];
	        this.entryTime = source["entryTime"];
	        this.exitTime = source["exitTime"];
	        this.entryPrice = source["entryPrice"];
	        this.exitPrice = source["exitPrice"];
	        this.stopLoss = source["stopLoss"];
	        this.takeProfit = source["takeProfit"];
	        this.quantity = source["quantity"];
	        this.orderValueUSDT = source["orderValueUSDT"];
	        this.pnl = source["pnl"];
	        this.pnlPercent = source["pnlPercent"];
	        this.balanceAfter = source["balanceAfter"];
	        this.exitReason = source["exitReason"];
	        this.holdBars = source["holdBars"];
	    }
	}
	export class BoxRangeReport {
	    strategyName: string;
	    dataPath: string;
	    resultPath: string;
	    generatedAt: string;
	    initialBalance: number;
	    positionSizeUSDT: number;
	    params: strategy.BoxRangeReversalParams;
	    klines: market.KLine[];
	    signals: strategy.Signal[];
	    trades: Trade[];
	    summary: Summary;
	
	    static createFrom(source: any = {}) {
	        return new BoxRangeReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategyName = source["strategyName"];
	        this.dataPath = source["dataPath"];
	        this.resultPath = source["resultPath"];
	        this.generatedAt = source["generatedAt"];
	        this.initialBalance = source["initialBalance"];
	        this.positionSizeUSDT = source["positionSizeUSDT"];
	        this.params = this.convertValues(source["params"], strategy.BoxRangeReversalParams);
	        this.klines = this.convertValues(source["klines"], market.KLine);
	        this.signals = this.convertValues(source["signals"], strategy.Signal);
	        this.trades = this.convertValues(source["trades"], Trade);
	        this.summary = this.convertValues(source["summary"], Summary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EMAReport {
	    strategyName: string;
	    dataPath: string;
	    resultPath: string;
	    generatedAt: string;
	    initialBalance: number;
	    positionSizeUSDT: number;
	    params: strategy.EMATrendPullbackParams;
	    klines: market.KLine[];
	    signals: strategy.Signal[];
	    trades: Trade[];
	    summary: Summary;
	
	    static createFrom(source: any = {}) {
	        return new EMAReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategyName = source["strategyName"];
	        this.dataPath = source["dataPath"];
	        this.resultPath = source["resultPath"];
	        this.generatedAt = source["generatedAt"];
	        this.initialBalance = source["initialBalance"];
	        this.positionSizeUSDT = source["positionSizeUSDT"];
	        this.params = this.convertValues(source["params"], strategy.EMATrendPullbackParams);
	        this.klines = this.convertValues(source["klines"], market.KLine);
	        this.signals = this.convertValues(source["signals"], strategy.Signal);
	        this.trades = this.convertValues(source["trades"], Trade);
	        this.summary = this.convertValues(source["summary"], Summary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Report {
	    strategyName: string;
	    dataPath: string;
	    resultPath: string;
	    generatedAt: string;
	    initialBalance: number;
	    positionSizeUSDT: number;
	    params: strategy.BoxPullbackParams;
	    klines: market.KLine[];
	    signals: strategy.Signal[];
	    trades: Trade[];
	    summary: Summary;
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategyName = source["strategyName"];
	        this.dataPath = source["dataPath"];
	        this.resultPath = source["resultPath"];
	        this.generatedAt = source["generatedAt"];
	        this.initialBalance = source["initialBalance"];
	        this.positionSizeUSDT = source["positionSizeUSDT"];
	        this.params = this.convertValues(source["params"], strategy.BoxPullbackParams);
	        this.klines = this.convertValues(source["klines"], market.KLine);
	        this.signals = this.convertValues(source["signals"], strategy.Signal);
	        this.trades = this.convertValues(source["trades"], Trade);
	        this.summary = this.convertValues(source["summary"], Summary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RunBoxRangeRequest {
	    dataPath: string;
	    strategyName: string;
	    params: strategy.BoxRangeReversalParams;
	    initialBalance: number;
	    positionSizeUSDT: number;
	    resultPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RunBoxRangeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataPath = source["dataPath"];
	        this.strategyName = source["strategyName"];
	        this.params = this.convertValues(source["params"], strategy.BoxRangeReversalParams);
	        this.initialBalance = source["initialBalance"];
	        this.positionSizeUSDT = source["positionSizeUSDT"];
	        this.resultPath = source["resultPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RunEMARequest {
	    dataPath: string;
	    strategyName: string;
	    params: strategy.EMATrendPullbackParams;
	    initialBalance: number;
	    positionSizeUSDT: number;
	    resultPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RunEMARequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataPath = source["dataPath"];
	        this.strategyName = source["strategyName"];
	        this.params = this.convertValues(source["params"], strategy.EMATrendPullbackParams);
	        this.initialBalance = source["initialBalance"];
	        this.positionSizeUSDT = source["positionSizeUSDT"];
	        this.resultPath = source["resultPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RunRequest {
	    dataPath: string;
	    strategyName: string;
	    params: strategy.BoxPullbackParams;
	    initialBalance: number;
	    positionSizeUSDT: number;
	    resultPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataPath = source["dataPath"];
	        this.strategyName = source["strategyName"];
	        this.params = this.convertValues(source["params"], strategy.BoxPullbackParams);
	        this.initialBalance = source["initialBalance"];
	        this.positionSizeUSDT = source["positionSizeUSDT"];
	        this.resultPath = source["resultPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace market {
	
	export class KLine {
	    openTime: number;
	    open: number;
	    high: number;
	    low: number;
	    close: number;
	    volume: number;
	    closeTime: number;
	    quoteAssetVolume: number;
	    numberOfTrades: number;
	    takerBuyVolume: number;
	    takerBuyQuoteVolume: number;
	
	    static createFrom(source: any = {}) {
	        return new KLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.openTime = source["openTime"];
	        this.open = source["open"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.close = source["close"];
	        this.volume = source["volume"];
	        this.closeTime = source["closeTime"];
	        this.quoteAssetVolume = source["quoteAssetVolume"];
	        this.numberOfTrades = source["numberOfTrades"];
	        this.takerBuyVolume = source["takerBuyVolume"];
	        this.takerBuyQuoteVolume = source["takerBuyQuoteVolume"];
	    }
	}

}

export namespace strategy {
	
	export class BoxPullbackParams {
	    lookaheadN: number;
	    minK1BodyPercent: number;
	    k1StrengthLookback: number;
	    minK1BodyToAvgRatio: number;
	    trendMAPeriod: number;
	    minBoxRangePercent: number;
	    maxBoxRangePercent: number;
	    touchTolerancePercent: number;
	    minConfirmWickBodyRatio: number;
	    cooldownBars: number;
	    riskRewardRatio: number;
	
	    static createFrom(source: any = {}) {
	        return new BoxPullbackParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lookaheadN = source["lookaheadN"];
	        this.minK1BodyPercent = source["minK1BodyPercent"];
	        this.k1StrengthLookback = source["k1StrengthLookback"];
	        this.minK1BodyToAvgRatio = source["minK1BodyToAvgRatio"];
	        this.trendMAPeriod = source["trendMAPeriod"];
	        this.minBoxRangePercent = source["minBoxRangePercent"];
	        this.maxBoxRangePercent = source["maxBoxRangePercent"];
	        this.touchTolerancePercent = source["touchTolerancePercent"];
	        this.minConfirmWickBodyRatio = source["minConfirmWickBodyRatio"];
	        this.cooldownBars = source["cooldownBars"];
	        this.riskRewardRatio = source["riskRewardRatio"];
	    }
	}
	export class BoxRangeReversalParams {
	    impulseLookback: number;
	    consolidationLookback: number;
	    atrPeriod: number;
	    minImpulsePercent: number;
	    minImpulseATRRatio: number;
	    minBoxWidthPercent: number;
	    maxBoxWidthPercent: number;
	    consolidationVolumeRatio: number;
	    consolidationATRRatio: number;
	    minBoundaryTouches: number;
	    edgeTolerancePercent: number;
	    minRejectWickBodyRatio: number;
	    stopATRMultiplier: number;
	    cooldownBars: number;
	    takeProfitFactor: number;
	
	    static createFrom(source: any = {}) {
	        return new BoxRangeReversalParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.impulseLookback = source["impulseLookback"];
	        this.consolidationLookback = source["consolidationLookback"];
	        this.atrPeriod = source["atrPeriod"];
	        this.minImpulsePercent = source["minImpulsePercent"];
	        this.minImpulseATRRatio = source["minImpulseATRRatio"];
	        this.minBoxWidthPercent = source["minBoxWidthPercent"];
	        this.maxBoxWidthPercent = source["maxBoxWidthPercent"];
	        this.consolidationVolumeRatio = source["consolidationVolumeRatio"];
	        this.consolidationATRRatio = source["consolidationATRRatio"];
	        this.minBoundaryTouches = source["minBoundaryTouches"];
	        this.edgeTolerancePercent = source["edgeTolerancePercent"];
	        this.minRejectWickBodyRatio = source["minRejectWickBodyRatio"];
	        this.stopATRMultiplier = source["stopATRMultiplier"];
	        this.cooldownBars = source["cooldownBars"];
	        this.takeProfitFactor = source["takeProfitFactor"];
	    }
	}
	export class EMATrendPullbackParams {
	    fastPeriod: number;
	    slowPeriod: number;
	    breakoutLookback: number;
	    pullbackLookahead: number;
	    pullbackTolerancePercent: number;
	    atrPeriod: number;
	    stopATRMultiplier: number;
	    cooldownBars: number;
	    riskRewardRatio: number;
	
	    static createFrom(source: any = {}) {
	        return new EMATrendPullbackParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fastPeriod = source["fastPeriod"];
	        this.slowPeriod = source["slowPeriod"];
	        this.breakoutLookback = source["breakoutLookback"];
	        this.pullbackLookahead = source["pullbackLookahead"];
	        this.pullbackTolerancePercent = source["pullbackTolerancePercent"];
	        this.atrPeriod = source["atrPeriod"];
	        this.stopATRMultiplier = source["stopATRMultiplier"];
	        this.cooldownBars = source["cooldownBars"];
	        this.riskRewardRatio = source["riskRewardRatio"];
	    }
	}
	export class Signal {
	    strategyName: string;
	    direction: string;
	    k1Index: number;
	    triggerIndex: number;
	    entryIndex: number;
	    k1OpenTime: number;
	    triggerTime: number;
	    entryTime: number;
	    boxHigh: number;
	    boxLow: number;
	    entryPrice: number;
	    stopLoss: number;
	    takeProfit: number;
	    riskRewardRatio: number;
	    confirmBarOpen: number;
	    confirmBarClose: number;
	    confirmBarLow: number;
	    confirmBarHigh: number;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new Signal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategyName = source["strategyName"];
	        this.direction = source["direction"];
	        this.k1Index = source["k1Index"];
	        this.triggerIndex = source["triggerIndex"];
	        this.entryIndex = source["entryIndex"];
	        this.k1OpenTime = source["k1OpenTime"];
	        this.triggerTime = source["triggerTime"];
	        this.entryTime = source["entryTime"];
	        this.boxHigh = source["boxHigh"];
	        this.boxLow = source["boxLow"];
	        this.entryPrice = source["entryPrice"];
	        this.stopLoss = source["stopLoss"];
	        this.takeProfit = source["takeProfit"];
	        this.riskRewardRatio = source["riskRewardRatio"];
	        this.confirmBarOpen = source["confirmBarOpen"];
	        this.confirmBarClose = source["confirmBarClose"];
	        this.confirmBarLow = source["confirmBarLow"];
	        this.confirmBarHigh = source["confirmBarHigh"];
	        this.reason = source["reason"];
	    }
	}

}

