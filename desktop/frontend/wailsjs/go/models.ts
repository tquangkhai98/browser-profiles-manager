export namespace main {
	
	export class BrowserItem {
	    name: string;
	    id: string;
	    exe_path: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.id = source["id"];
	        this.exe_path = source["exe_path"];
	    }
	}
	export class CredentialSite {
	    domain: string;
	    cookies: number;
	    logins: number;
	
	    static createFrom(source: any = {}) {
	        return new CredentialSite(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.cookies = source["cookies"];
	        this.logins = source["logins"];
	    }
	}
	export class CredentialResult {
	    profile_name: string;
	    sites: CredentialSite[];
	    total_cookies: number;
	    total_logins: number;
	
	    static createFrom(source: any = {}) {
	        return new CredentialResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profile_name = source["profile_name"];
	        this.sites = this.convertValues(source["sites"], CredentialSite);
	        this.total_cookies = source["total_cookies"];
	        this.total_logins = source["total_logins"];
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
	
	export class ProfileInfo {
	    name: string;
	    browser: string;
	    data_dir: string;
	    created_at: string;
	    last_used: string;
	    locked: boolean;
	    lock_pid: number;
	    lock_by: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.browser = source["browser"];
	        this.data_dir = source["data_dir"];
	        this.created_at = source["created_at"];
	        this.last_used = source["last_used"];
	        this.locked = source["locked"];
	        this.lock_pid = source["lock_pid"];
	        this.lock_by = source["lock_by"];
	    }
	}
	export class SyncResult {
	    files_copied: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files_copied = source["files_copied"];
	        this.message = source["message"];
	    }
	}

}

