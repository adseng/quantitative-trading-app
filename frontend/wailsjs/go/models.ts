export namespace cases {
	
	export class TestCase {
	    id: number;
	    name: string;
	    useMA: boolean;
	    maShort: number;
	    maLong: number;
	    maWeight: number;
	    useTrend: boolean;
	    trendN: number;
	    trendWeight: number;
	    useRSI: boolean;
	    rsiPeriod: number;
	    rsiOverbought: number;
	    rsiOversold: number;
	    rsiWeight: number;
	    useMACD: boolean;
	    macdFast: number;
	    macdSlow: number;
	    macdSignal: number;
	    macdWeight: number;
	    useBoll: boolean;
	    bollPeriod: number;
	    bollMultiplier: number;
	    bollWeight: number;
	    useBreakout: boolean;
	    breakoutPeriod: number;
	    breakoutWeight: number;
	    usePriceVsMA: boolean;
	    priceVsMAPeriod: number;
	    priceVsMAWeight: number;
	    useATR: boolean;
	    atrPeriod: number;
	    atrWeight: number;
	    useVolume: boolean;
	    volumePeriod: number;
	    volumeWeight: number;
	    useSession: boolean;
	    sessionWeight: number;
	    useMACross: boolean;
	    macrossShort: number;
	    macrossLong: number;
	    macrossWeight: number;
	    macrossWindow: number;
	    macrossPreempt: number;
	
	    static createFrom(source: any = {}) {
	        return new TestCase(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.useMA = source["useMA"];
	        this.maShort = source["maShort"];
	        this.maLong = source["maLong"];
	        this.maWeight = source["maWeight"];
	        this.useTrend = source["useTrend"];
	        this.trendN = source["trendN"];
	        this.trendWeight = source["trendWeight"];
	        this.useRSI = source["useRSI"];
	        this.rsiPeriod = source["rsiPeriod"];
	        this.rsiOverbought = source["rsiOverbought"];
	        this.rsiOversold = source["rsiOversold"];
	        this.rsiWeight = source["rsiWeight"];
	        this.useMACD = source["useMACD"];
	        this.macdFast = source["macdFast"];
	        this.macdSlow = source["macdSlow"];
	        this.macdSignal = source["macdSignal"];
	        this.macdWeight = source["macdWeight"];
	        this.useBoll = source["useBoll"];
	        this.bollPeriod = source["bollPeriod"];
	        this.bollMultiplier = source["bollMultiplier"];
	        this.bollWeight = source["bollWeight"];
	        this.useBreakout = source["useBreakout"];
	        this.breakoutPeriod = source["breakoutPeriod"];
	        this.breakoutWeight = source["breakoutWeight"];
	        this.usePriceVsMA = source["usePriceVsMA"];
	        this.priceVsMAPeriod = source["priceVsMAPeriod"];
	        this.priceVsMAWeight = source["priceVsMAWeight"];
	        this.useATR = source["useATR"];
	        this.atrPeriod = source["atrPeriod"];
	        this.atrWeight = source["atrWeight"];
	        this.useVolume = source["useVolume"];
	        this.volumePeriod = source["volumePeriod"];
	        this.volumeWeight = source["volumeWeight"];
	        this.useSession = source["useSession"];
	        this.sessionWeight = source["sessionWeight"];
	        this.useMACross = source["useMACross"];
	        this.macrossShort = source["macrossShort"];
	        this.macrossLong = source["macrossLong"];
	        this.macrossWeight = source["macrossWeight"];
	        this.macrossWindow = source["macrossWindow"];
	        this.macrossPreempt = source["macrossPreempt"];
	    }
	}
	export class TestResult {
	    testCase: TestCase;
	    accuracy: number;
	    correct: number;
	    total: number;
	    signalCount: number;
	    signalAccuracy: number;
	    avgScore: number;
	    avgAbsScore: number;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.testCase = this.convertValues(source["testCase"], TestCase);
	        this.accuracy = source["accuracy"];
	        this.correct = source["correct"];
	        this.total = source["total"];
	        this.signalCount = source["signalCount"];
	        this.signalAccuracy = source["signalAccuracy"];
	        this.avgScore = source["avgScore"];
	        this.avgAbsScore = source["avgAbsScore"];
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

export namespace coze {
	
	export class CozeScenario {
	    direction: string;
	    probability: number;
	    setup_logic: string;
	    trigger_condition: string;
	    entry_price?: number;
	    stop_loss?: number;
	    take_profit_1?: number;
	    take_profit_2?: number;
	    risk_reward_ratio?: number;
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new CozeScenario(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.direction = source["direction"];
	        this.probability = source["probability"];
	        this.setup_logic = source["setup_logic"];
	        this.trigger_condition = source["trigger_condition"];
	        this.entry_price = source["entry_price"];
	        this.stop_loss = source["stop_loss"];
	        this.take_profit_1 = source["take_profit_1"];
	        this.take_profit_2 = source["take_profit_2"];
	        this.risk_reward_ratio = source["risk_reward_ratio"];
	        this.action = source["action"];
	    }
	}
	export class CozeStructuredResult {
	    timestamp: string;
	    symbol: string;
	    current_price: number;
	    market_structure: string;
	    scenarios: CozeScenario[];
	
	    static createFrom(source: any = {}) {
	        return new CozeStructuredResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.symbol = source["symbol"];
	        this.current_price = source["current_price"];
	        this.market_structure = source["market_structure"];
	        this.scenarios = this.convertValues(source["scenarios"], CozeScenario);
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

export namespace factor {
	
	export class FactorDetail {
	    maScore: number;
	    trendContrib: number;
	    rsiContrib: number;
	    macdContrib: number;
	    bollContrib: number;
	    breakoutContrib: number;
	    priceVsMAContrib: number;
	    atrContrib: number;
	    volumeContrib: number;
	    sessionContrib: number;
	    macrossContrib: number;
	    bullScore: number;
	    bearScore: number;
	
	    static createFrom(source: any = {}) {
	        return new FactorDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maScore = source["maScore"];
	        this.trendContrib = source["trendContrib"];
	        this.rsiContrib = source["rsiContrib"];
	        this.macdContrib = source["macdContrib"];
	        this.bollContrib = source["bollContrib"];
	        this.breakoutContrib = source["breakoutContrib"];
	        this.priceVsMAContrib = source["priceVsMAContrib"];
	        this.atrContrib = source["atrContrib"];
	        this.volumeContrib = source["volumeContrib"];
	        this.sessionContrib = source["sessionContrib"];
	        this.macrossContrib = source["macrossContrib"];
	        this.bullScore = source["bullScore"];
	        this.bearScore = source["bearScore"];
	    }
	}
	export class BacktestResult {
	    index: number;
	    openTime: number;
	    actual: number;
	    predicted: number;
	    correct: boolean;
	    factors?: FactorDetail;
	
	    static createFrom(source: any = {}) {
	        return new BacktestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.openTime = source["openTime"];
	        this.actual = source["actual"];
	        this.predicted = source["predicted"];
	        this.correct = source["correct"];
	        this.factors = this.convertValues(source["factors"], FactorDetail);
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
	export class BacktestResultSummary {
	    results: BacktestResult[];
	    total: number;
	    correct: number;
	    accuracy: number;
	    signalCount: number;
	    signalAccuracy: number;
	    avgScore: number;
	    avgAbsScore: number;
	
	    static createFrom(source: any = {}) {
	        return new BacktestResultSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.results = this.convertValues(source["results"], BacktestResult);
	        this.total = source["total"];
	        this.correct = source["correct"];
	        this.accuracy = source["accuracy"];
	        this.signalCount = source["signalCount"];
	        this.signalAccuracy = source["signalAccuracy"];
	        this.avgScore = source["avgScore"];
	        this.avgAbsScore = source["avgAbsScore"];
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

