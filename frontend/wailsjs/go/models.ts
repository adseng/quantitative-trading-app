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

