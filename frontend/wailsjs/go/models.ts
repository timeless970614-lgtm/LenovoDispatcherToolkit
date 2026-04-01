export namespace backend {
	
	export class DispatcherFeature {
	    bit: number;
	    name: string;
	    desc: string;
	    enabled: boolean;
	    group: string;
	
	    static createFrom(source: any = {}) {
	        return new DispatcherFeature(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bit = source["bit"];
	        this.name = source["name"];
	        this.desc = source["desc"];
	        this.enabled = source["enabled"];
	        this.group = source["group"];
	    }
	}
	export class DYTCInfo {
	    currentMode: string;
	    currentDispatcherMode: string;
	    dccCapability: number;
	    geekCapability: number;
	    aiEngineMode: string;
	    dispatcherFunction: number;
	    dispatcherThreshold: number;
	    enableFunc: number;
	    dispatcherFeatures: DispatcherFeature[];
	
	    static createFrom(source: any = {}) {
	        return new DYTCInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentMode = source["currentMode"];
	        this.currentDispatcherMode = source["currentDispatcherMode"];
	        this.dccCapability = source["dccCapability"];
	        this.geekCapability = source["geekCapability"];
	        this.aiEngineMode = source["aiEngineMode"];
	        this.dispatcherFunction = source["dispatcherFunction"];
	        this.dispatcherThreshold = source["dispatcherThreshold"];
	        this.enableFunc = source["enableFunc"];
	        this.dispatcherFeatures = this.convertValues(source["dispatcherFeatures"], DispatcherFeature);
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
	
	export class DispatcherInfo {
	    driverVersion: string;
	    description: string;
	    currentMode: string;
	    aiEngineMode: string;
	    autoMode: number;
	
	    static createFrom(source: any = {}) {
	        return new DispatcherInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.driverVersion = source["driverVersion"];
	        this.description = source["description"];
	        this.currentMode = source["currentMode"];
	        this.aiEngineMode = source["aiEngineMode"];
	        this.autoMode = source["autoMode"];
	    }
	}
	export class EnableFuncPolicy {
	    bit: number;
	    name: string;
	    desc: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EnableFuncPolicy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bit = source["bit"];
	        this.name = source["name"];
	        this.desc = source["desc"];
	        this.enabled = source["enabled"];
	    }
	}
	export class GPUInfo {
	    name: string;
	    vendorId: number;
	    deviceId: number;
	    subVendorId: number;
	    subSystemId: number;
	    revisionId: number;
	    driverVersion: string;
	    driverDate: string;
	    dedicatedMemory: number;
	    sharedMemory: number;
	    totalMemory: number;
	    isDiscrete: boolean;
	    hardwareId: string;
	    busNumber: number;
	
	    static createFrom(source: any = {}) {
	        return new GPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.vendorId = source["vendorId"];
	        this.deviceId = source["deviceId"];
	        this.subVendorId = source["subVendorId"];
	        this.subSystemId = source["subSystemId"];
	        this.revisionId = source["revisionId"];
	        this.driverVersion = source["driverVersion"];
	        this.driverDate = source["driverDate"];
	        this.dedicatedMemory = source["dedicatedMemory"];
	        this.sharedMemory = source["sharedMemory"];
	        this.totalMemory = source["totalMemory"];
	        this.isDiscrete = source["isDiscrete"];
	        this.hardwareId = source["hardwareId"];
	        this.busNumber = source["busNumber"];
	    }
	}
	export class GPUProcess {
	    pid: number;
	    name: string;
	    memory: string;
	
	    static createFrom(source: any = {}) {
	        return new GPUProcess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.memory = source["memory"];
	    }
	}
	export class IGPUStatus {
	    available: boolean;
	    mode: number;
	
	    static createFrom(source: any = {}) {
	        return new IGPUStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.mode = source["mode"];
	    }
	}
	export class LogFileInfo {
	    name: string;
	    size: number;
	    modTime: string;
	
	    static createFrom(source: any = {}) {
	        return new LogFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	    }
	}
	export class MLLogStatus {
	    isCapturing: boolean;
	    startTime: string;
	    eventCount: number;
	    outputFile: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new MLLogStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isCapturing = source["isCapturing"];
	        this.startTime = source["startTime"];
	        this.eventCount = source["eventCount"];
	        this.outputFile = source["outputFile"];
	        this.error = source["error"];
	    }
	}
	export class ModeCheckFeature {
	    name: string;
	    value: string;
	    support: string;
	
	    static createFrom(source: any = {}) {
	        return new ModeCheckFeature(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.support = source["support"];
	    }
	}
	export class ModeCheckInfo {
	    model: string;
	    biosVersion: string;
	    memoryType: string;
	    driverVersion: string;
	    dispatcherMode: string;
	    dispatcherVersion: string;
	    aiEngineMode: string;
	    mainVer: string;
	    dytcValue: number;
	    dytcBinary: string;
	    enableFuncValue: number;
	    enableFuncHex: string;
	    enableFuncPolicies: EnableFuncPolicy[];
	    features: ModeCheckFeature[];
	
	    static createFrom(source: any = {}) {
	        return new ModeCheckInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.biosVersion = source["biosVersion"];
	        this.memoryType = source["memoryType"];
	        this.driverVersion = source["driverVersion"];
	        this.dispatcherMode = source["dispatcherMode"];
	        this.dispatcherVersion = source["dispatcherVersion"];
	        this.aiEngineMode = source["aiEngineMode"];
	        this.mainVer = source["mainVer"];
	        this.dytcValue = source["dytcValue"];
	        this.dytcBinary = source["dytcBinary"];
	        this.enableFuncValue = source["enableFuncValue"];
	        this.enableFuncHex = source["enableFuncHex"];
	        this.enableFuncPolicies = this.convertValues(source["enableFuncPolicies"], EnableFuncPolicy);
	        this.features = this.convertValues(source["features"], ModeCheckFeature);
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
	export class NVIDIAStatus {
	    detected: boolean;
	    nvmlLoaded: boolean;
	    serviceRunning: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NVIDIAStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.detected = source["detected"];
	        this.nvmlLoaded = source["nvmlLoaded"];
	        this.serviceRunning = source["serviceRunning"];
	    }
	}
	export class PPMSetting {
	    name: string;
	    guid: string;
	    acValue: number;
	    dcValue: number;
	    found: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PPMSetting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.guid = source["guid"];
	        this.acValue = source["acValue"];
	        this.dcValue = source["dcValue"];
	        this.found = source["found"];
	    }
	}
	export class PPMSettings {
	    epp: PPMSetting;
	    epp1: PPMSetting;
	    heteroInc: PPMSetting;
	    heteroDec: PPMSetting;
	    maxFreq: PPMSetting;
	    maxFreq1: PPMSetting;
	    softPark: PPMSetting;
	    cpMinCores: PPMSetting;
	    minPerf: PPMSetting;
	    maxPerf: PPMSetting;
	    schemeName: string;
	    schemeGUID: string;
	
	    static createFrom(source: any = {}) {
	        return new PPMSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.epp = this.convertValues(source["epp"], PPMSetting);
	        this.epp1 = this.convertValues(source["epp1"], PPMSetting);
	        this.heteroInc = this.convertValues(source["heteroInc"], PPMSetting);
	        this.heteroDec = this.convertValues(source["heteroDec"], PPMSetting);
	        this.maxFreq = this.convertValues(source["maxFreq"], PPMSetting);
	        this.maxFreq1 = this.convertValues(source["maxFreq1"], PPMSetting);
	        this.softPark = this.convertValues(source["softPark"], PPMSetting);
	        this.cpMinCores = this.convertValues(source["cpMinCores"], PPMSetting);
	        this.minPerf = this.convertValues(source["minPerf"], PPMSetting);
	        this.maxPerf = this.convertValues(source["maxPerf"], PPMSetting);
	        this.schemeName = source["schemeName"];
	        this.schemeGUID = source["schemeGUID"];
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
	export class SSDInfo {
	    driveIndex: number;
	    name: string;
	    model: string;
	    serialNumber: string;
	    capacityBytes: number;
	    capacityStr: string;
	    protocol: string;
	    multiModeCapable: boolean;
	    currentMode: number;
	    currentModeStr: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SSDInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.driveIndex = source["driveIndex"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.serialNumber = source["serialNumber"];
	        this.capacityBytes = source["capacityBytes"];
	        this.capacityStr = source["capacityStr"];
	        this.protocol = source["protocol"];
	        this.multiModeCapable = source["multiModeCapable"];
	        this.currentMode = source["currentMode"];
	        this.currentModeStr = source["currentModeStr"];
	        this.error = source["error"];
	    }
	}
	export class SSDModeResult {
	    driveIndex: number;
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SSDModeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.driveIndex = source["driveIndex"];
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class SetResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SetResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class SystemInfo {
	    cpuName: string;
	    biosVersion: string;
	    osCaption: string;
	    osVersion: string;
	    totalMemoryGB: number;
	    memoryType: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpuName = source["cpuName"];
	        this.biosVersion = source["biosVersion"];
	        this.osCaption = source["osCaption"];
	        this.osVersion = source["osVersion"];
	        this.totalMemoryGB = source["totalMemoryGB"];
	        this.memoryType = source["memoryType"];
	    }
	}

}

