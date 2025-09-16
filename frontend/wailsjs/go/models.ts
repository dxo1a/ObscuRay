export namespace backend {
	
	export class Profile {
	    id: string;
	    name: string;
	    vless: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.vless = source["vless"];
	        this.isActive = source["isActive"];
	    }
	}

}

