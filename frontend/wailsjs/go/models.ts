export namespace backend {
	
	export class BlockRequest {
	    type: string;
	    target: string;
	    interface?: string;
	    pid?: number;
	
	    static createFrom(source: any = {}) {
	        return new BlockRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.target = source["target"];
	        this.interface = source["interface"];
	        this.pid = source["pid"];
	    }
	}
	export class BlockedEntry {
	    id: string;
	    type: string;
	    target: string;
	    interface?: string;
	    rule_name: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BlockedEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.target = source["target"];
	        this.interface = source["interface"];
	        this.rule_name = source["rule_name"];
	        this.active = source["active"];
	    }
	}
	export class BlockResult {
	    success: boolean;
	    message: string;
	    action: string;
	    entry?: BlockedEntry;
	    is_admin: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BlockResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.action = source["action"];
	        this.entry = this.convertValues(source["entry"], BlockedEntry);
	        this.is_admin = source["is_admin"];
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
	
	export class ConnectionInfo {
	    fd: number;
	    family: number;
	    type: number;
	    local_ip: string;
	    local_port: number;
	    remote_ip: string;
	    remote_port: number;
	    status: string;
	    pid: number;
	    process_name: string;
	    protocol: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fd = source["fd"];
	        this.family = source["family"];
	        this.type = source["type"];
	        this.local_ip = source["local_ip"];
	        this.local_port = source["local_port"];
	        this.remote_ip = source["remote_ip"];
	        this.remote_port = source["remote_port"];
	        this.status = source["status"];
	        this.pid = source["pid"];
	        this.process_name = source["process_name"];
	        this.protocol = source["protocol"];
	    }
	}
	export class IPAddressInfo {
	    address: string;
	    network: string;
	    address_type: string;
	
	    static createFrom(source: any = {}) {
	        return new IPAddressInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.network = source["network"];
	        this.address_type = source["address_type"];
	    }
	}
	export class InterfaceInfo {
	    name: string;
	    description: string;
	    mac_address: string;
	    is_up: boolean;
	    is_loopback: boolean;
	    mtu: number;
	    flags: string[];
	    ip_addresses: IPAddressInfo[];
	    connections: ConnectionInfo[];
	
	    static createFrom(source: any = {}) {
	        return new InterfaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.mac_address = source["mac_address"];
	        this.is_up = source["is_up"];
	        this.is_loopback = source["is_loopback"];
	        this.mtu = source["mtu"];
	        this.flags = source["flags"];
	        this.ip_addresses = this.convertValues(source["ip_addresses"], IPAddressInfo);
	        this.connections = this.convertValues(source["connections"], ConnectionInfo);
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
	export class ProcessDetail {
	    pid: number;
	    name: string;
	    exe_path: string;
	    connection_count: number;
	    rx_bytes: number;
	    tx_bytes: number;
	    rx_bps: number;
	    tx_bps: number;
	    rx_bps_str: string;
	    tx_bps_str: string;
	    estimated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProcessDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.exe_path = source["exe_path"];
	        this.connection_count = source["connection_count"];
	        this.rx_bytes = source["rx_bytes"];
	        this.tx_bytes = source["tx_bytes"];
	        this.rx_bps = source["rx_bps"];
	        this.tx_bps = source["tx_bps"];
	        this.rx_bps_str = source["rx_bps_str"];
	        this.tx_bps_str = source["tx_bps_str"];
	        this.estimated = source["estimated"];
	    }
	}
	export class ProcessNetIO {
	    pid: number;
	    bytes_recv: number;
	    bytes_sent: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessNetIO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.bytes_recv = source["bytes_recv"];
	        this.bytes_sent = source["bytes_sent"];
	    }
	}
	export class ScanOptions {
	    include_loopback: boolean;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.include_loopback = source["include_loopback"];
	        this.timeout = source["timeout"];
	    }
	}
	export class ScanResult {
	    interfaces: InterfaceInfo[];
	    total_interfaces: number;
	    total_connections: number;
	    timestamp: number;
	    duration_ms: number;
	    error?: string;
	    permission_error: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interfaces = this.convertValues(source["interfaces"], InterfaceInfo);
	        this.total_interfaces = source["total_interfaces"];
	        this.total_connections = source["total_connections"];
	        this.timestamp = source["timestamp"];
	        this.duration_ms = source["duration_ms"];
	        this.error = source["error"];
	        this.permission_error = source["permission_error"];
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
	export class ThroughputData {
	    interface: string;
	    rx_bytes: number;
	    tx_bytes: number;
	    rx_pkts: number;
	    tx_pkts: number;
	    rx_bps: number;
	    tx_bps: number;
	    rx_bps_str: string;
	    tx_bps_str: string;
	
	    static createFrom(source: any = {}) {
	        return new ThroughputData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interface = source["interface"];
	        this.rx_bytes = source["rx_bytes"];
	        this.tx_bytes = source["tx_bytes"];
	        this.rx_pkts = source["rx_pkts"];
	        this.tx_pkts = source["tx_pkts"];
	        this.rx_bps = source["rx_bps"];
	        this.tx_bps = source["tx_bps"];
	        this.rx_bps_str = source["rx_bps_str"];
	        this.tx_bps_str = source["tx_bps_str"];
	    }
	}

}

