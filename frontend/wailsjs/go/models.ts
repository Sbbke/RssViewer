export namespace dto {
	
	export class BriefingTextResponse {
	    body: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BriefingTextResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.body = source["body"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class PostItem {
	    id: number;
	    title: string;
	    content: string;
	    publishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new PostItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.publishedAt = source["publishedAt"];
	    }
	}
	export class RssItem {
	    id: number;
	    title: string;
	    // Go type: time
	    subscribedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RssItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.subscribedAt = this.convertValues(source["subscribedAt"], null);
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
	export class RssResponse {
	    info: RssItem;
	    posts: PostItem[];
	
	    static createFrom(source: any = {}) {
	        return new RssResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.info = this.convertValues(source["info"], RssItem);
	        this.posts = this.convertValues(source["posts"], PostItem);
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
	export class TopicResponse {
	    topicId: number;
	    rss: RssItem[];
	    name: string;
	    summary?: BriefingTextResponse;
	    summaryId: number;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TopicResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.topicId = source["topicId"];
	        this.rss = this.convertValues(source["rss"], RssItem);
	        this.name = source["name"];
	        this.summary = this.convertValues(source["summary"], BriefingTextResponse);
	        this.summaryId = source["summaryId"];
	        this.createdAt = source["createdAt"];
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

