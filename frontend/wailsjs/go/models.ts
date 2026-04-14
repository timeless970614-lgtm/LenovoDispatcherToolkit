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
	export class DynamicLogResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DynamicLogResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class EPOTStatus {
	    epot: number;
	    epp: number;
	    epp1: number;
	    ppmFrequencyLimit: number;
	    ppmFrequencyLimit1: number;
	    ppmCpMin: number;
	    ppmCpMax: number;
	    softParking: number;
	
	    static createFrom(source: any = {}) {
	        return new EPOTStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.epot = source["epot"];
	        this.epp = source["epp"];
	        this.epp1 = source["epp1"];
	        this.ppmFrequencyLimit = source["ppmFrequencyLimit"];
	        this.ppmFrequencyLimit1 = source["ppmFrequencyLimit1"];
	        this.ppmCpMin = source["ppmCpMin"];
	        this.ppmCpMax = source["ppmCpMax"];
	        this.softParking = source["softParking"];
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
	export class GPUAutoGear {
	    available: boolean;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new GPUAutoGear(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.value = source["value"];
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
	export class GPUPrefStatus {
	    available: boolean;
	    value: number;
	    label: string;
	    pcmStatus: number;
	    pcmStatusAvail: boolean;
	    pcmLabel: string;
	
	    static createFrom(source: any = {}) {
	        return new GPUPrefStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.value = source["value"];
	        this.label = source["label"];
	        this.pcmStatus = source["pcmStatus"];
	        this.pcmStatusAvail = source["pcmStatusAvail"];
	        this.pcmLabel = source["pcmLabel"];
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
	export class IntelGPUFreqTestResult {
	    success: boolean;
	    message: string;
	    minFreq: number;
	    maxFreq: number;
	
	    static createFrom(source: any = {}) {
	        return new IntelGPUFreqTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.minFreq = source["minFreq"];
	        this.maxFreq = source["maxFreq"];
	    }
	}
	export class IntelGPUFrequency {
	    available: boolean;
	    minFreq: number;
	    maxFreq: number;
	    currentMin: number;
	    currentMax: number;
	    requestedMHz: number;
	    actualMHz: number;
	    tdpMHz: number;
	    efficientMHz: number;
	    gpuUtilization: number;
	    gpuName: string;
	    driverVersion: string;
	    driverDate: string;
	    minDriverVersion: string;
	    driverOK: boolean;
	    adapterIndex: number;
	    regKeyPath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new IntelGPUFrequency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.minFreq = source["minFreq"];
	        this.maxFreq = source["maxFreq"];
	        this.currentMin = source["currentMin"];
	        this.currentMax = source["currentMax"];
	        this.requestedMHz = source["requestedMHz"];
	        this.actualMHz = source["actualMHz"];
	        this.tdpMHz = source["tdpMHz"];
	        this.efficientMHz = source["efficientMHz"];
	        this.gpuUtilization = source["gpuUtilization"];
	        this.gpuName = source["gpuName"];
	        this.driverVersion = source["driverVersion"];
	        this.driverDate = source["driverDate"];
	        this.minDriverVersion = source["minDriverVersion"];
	        this.driverOK = source["driverOK"];
	        this.adapterIndex = source["adapterIndex"];
	        this.regKeyPath = source["regKeyPath"];
	        this.error = source["error"];
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
	    outputCSV: string;
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
	        this.outputCSV = source["outputCSV"];
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
	export class NPUCTCPHYInfo {
	    groupId: number;
	    chipId: number;
	
	    static createFrom(source: any = {}) {
	        return new NPUCTCPHYInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.chipId = source["chipId"];
	    }
	}
	export class NPUDeviceInfo {
	    numDevices: number;
	    deviceIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new NPUDeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.numDevices = source["numDevices"];
	        this.deviceIds = source["deviceIds"];
	    }
	}
	export class NPUDeviceMetrics {
	    ipuUtiliRate: number;
	    ipuVoltageMV: number;
	    ipuFrequencyHz: number;
	    boardPowerW: number;
	    temperatureC: number;
	    memTotalMB: number;
	    memUsedMB: number;
	    memAvailMB: number;
	    coreUtiliPct: number[];
	
	    static createFrom(source: any = {}) {
	        return new NPUDeviceMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ipuUtiliRate = source["ipuUtiliRate"];
	        this.ipuVoltageMV = source["ipuVoltageMV"];
	        this.ipuFrequencyHz = source["ipuFrequencyHz"];
	        this.boardPowerW = source["boardPowerW"];
	        this.temperatureC = source["temperatureC"];
	        this.memTotalMB = source["memTotalMB"];
	        this.memUsedMB = source["memUsedMB"];
	        this.memAvailMB = source["memAvailMB"];
	        this.coreUtiliPct = source["coreUtiliPct"];
	    }
	}
	export class NPUDeviceOverview {
	    devId: number;
	    devName: string;
	    vendorId: number;
	    serial: string;
	    computingPower: number;
	    coreCount: number;
	    ddrSizeMB: number;
	    dvfsMode: string;
	    dvfsModeDesc: string;
	    ipuUtiliPct: number;
	    ipuFreqGHz: number;
	    ipuVoltageMV: number;
	    temperatureC: number;
	    boardPowerW: number;
	    memTotalMB: number;
	    memUsedMB: number;
	    memAvailMB: number;
	    coreUtiliPct: number[];
	    sdkVersion: string;
	    driverVersion: string;
	    firmwareVer: string;
	
	    static createFrom(source: any = {}) {
	        return new NPUDeviceOverview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.devId = source["devId"];
	        this.devName = source["devName"];
	        this.vendorId = source["vendorId"];
	        this.serial = source["serial"];
	        this.computingPower = source["computingPower"];
	        this.coreCount = source["coreCount"];
	        this.ddrSizeMB = source["ddrSizeMB"];
	        this.dvfsMode = source["dvfsMode"];
	        this.dvfsModeDesc = source["dvfsModeDesc"];
	        this.ipuUtiliPct = source["ipuUtiliPct"];
	        this.ipuFreqGHz = source["ipuFreqGHz"];
	        this.ipuVoltageMV = source["ipuVoltageMV"];
	        this.temperatureC = source["temperatureC"];
	        this.boardPowerW = source["boardPowerW"];
	        this.memTotalMB = source["memTotalMB"];
	        this.memUsedMB = source["memUsedMB"];
	        this.memAvailMB = source["memAvailMB"];
	        this.coreUtiliPct = source["coreUtiliPct"];
	        this.sdkVersion = source["sdkVersion"];
	        this.driverVersion = source["driverVersion"];
	        this.firmwareVer = source["firmwareVer"];
	    }
	}
	export class NPUDeviceProperties {
	    vendorId: number;
	    serialNumber: string;
	    modelName: string;
	    computingPowerTOPS: number;
	    coreCount: number;
	    ddrSizeBytes: number;
	    ddrSizeMB: number;
	    firmwareVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new NPUDeviceProperties(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vendorId = source["vendorId"];
	        this.serialNumber = source["serialNumber"];
	        this.modelName = source["modelName"];
	        this.computingPowerTOPS = source["computingPowerTOPS"];
	        this.coreCount = source["coreCount"];
	        this.ddrSizeBytes = source["ddrSizeBytes"];
	        this.ddrSizeMB = source["ddrSizeMB"];
	        this.firmwareVersion = source["firmwareVersion"];
	    }
	}
	export class NPUPCIeInfo {
	    bdf: string;
	    bandwidth: string;
	
	    static createFrom(source: any = {}) {
	        return new NPUPCIeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bdf = source["bdf"];
	        this.bandwidth = source["bandwidth"];
	    }
	}
	export class NPUDeviceReport {
	    index: number;
	    properties: NPUDeviceProperties;
	    metrics: NPUDeviceMetrics;
	    pcieInfo: NPUPCIeInfo;
	    dvfsMode: string;
	    dvfsModeDesc: string;
	    ctcPhyInfo: NPUCTCPHYInfo;
	
	    static createFrom(source: any = {}) {
	        return new NPUDeviceReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.properties = this.convertValues(source["properties"], NPUDeviceProperties);
	        this.metrics = this.convertValues(source["metrics"], NPUDeviceMetrics);
	        this.pcieInfo = this.convertValues(source["pcieInfo"], NPUPCIeInfo);
	        this.dvfsMode = source["dvfsMode"];
	        this.dvfsModeDesc = source["dvfsModeDesc"];
	        this.ctcPhyInfo = this.convertValues(source["ctcPhyInfo"], NPUCTCPHYInfo);
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
	export class NPUSDKInfo {
	    buildtime: string;
	    sdkVersion: string;
	    driverVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new NPUSDKInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.buildtime = source["buildtime"];
	        this.sdkVersion = source["sdkVersion"];
	        this.driverVersion = source["driverVersion"];
	    }
	}
	export class NPUFullReport {
	    deviceCount: number;
	    sdkInfo: NPUSDKInfo;
	    devices: NPUDeviceReport[];
	
	    static createFrom(source: any = {}) {
	        return new NPUFullReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deviceCount = source["deviceCount"];
	        this.sdkInfo = this.convertValues(source["sdkInfo"], NPUSDKInfo);
	        this.devices = this.convertValues(source["devices"], NPUDeviceReport);
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
	
	export class NPUPowerAction {
	    success: boolean;
	    message: string;
	    newMode?: string;
	    newMaxMHz?: number;
	    newMinMHz?: number;
	
	    static createFrom(source: any = {}) {
	        return new NPUPowerAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.newMode = source["newMode"];
	        this.newMaxMHz = source["newMaxMHz"];
	        this.newMinMHz = source["newMinMHz"];
	    }
	}
	export class NPUPowerStatus {
	    dvfsMode: string;
	    curIpuFreqMHz: number;
	    lockIpuMaxMHz: number;
	    lockIpuMinMHz: number;
	    ipuLoadPct: number;
	    boardPowerW: number;
	    ddrTotalMB: number;
	    ddrFreeMB: number;
	    coreNum: number;
	    coreFreqMHz: number;
	    coreVoltageMV: number;
	    core0UtilPct: number;
	    core1UtilPct: number;
	    avgUtilPct: number;
	    ddr0TempC: number;
	    ddr2TempC: number;
	    ddr4TempC: number;
	    ddr5TempC: number;
	    core0TempC: number;
	    core1TempC: number;
	
	    static createFrom(source: any = {}) {
	        return new NPUPowerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dvfsMode = source["dvfsMode"];
	        this.curIpuFreqMHz = source["curIpuFreqMHz"];
	        this.lockIpuMaxMHz = source["lockIpuMaxMHz"];
	        this.lockIpuMinMHz = source["lockIpuMinMHz"];
	        this.ipuLoadPct = source["ipuLoadPct"];
	        this.boardPowerW = source["boardPowerW"];
	        this.ddrTotalMB = source["ddrTotalMB"];
	        this.ddrFreeMB = source["ddrFreeMB"];
	        this.coreNum = source["coreNum"];
	        this.coreFreqMHz = source["coreFreqMHz"];
	        this.coreVoltageMV = source["coreVoltageMV"];
	        this.core0UtilPct = source["core0UtilPct"];
	        this.core1UtilPct = source["core1UtilPct"];
	        this.avgUtilPct = source["avgUtilPct"];
	        this.ddr0TempC = source["ddr0TempC"];
	        this.ddr2TempC = source["ddr2TempC"];
	        this.ddr4TempC = source["ddr4TempC"];
	        this.ddr5TempC = source["ddr5TempC"];
	        this.core0TempC = source["core0TempC"];
	        this.core1TempC = source["core1TempC"];
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
	    CPUName: string;
	    CodeName: string;
	    BIOSVersion: string;
	    OSCaption: string;
	    OSVersion: string;
	    TotalMemoryGB: number;
	    MemoryType: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CPUName = source["CPUName"];
	        this.CodeName = source["CodeName"];
	        this.BIOSVersion = source["BIOSVersion"];
	        this.OSCaption = source["OSCaption"];
	        this.OSVersion = source["OSVersion"];
	        this.TotalMemoryGB = source["TotalMemoryGB"];
	        this.MemoryType = source["MemoryType"];
	    }
	}
	export class UninstallResult {
	    success: boolean;
	    message: string;
	    driversRemoved: number;
	
	    static createFrom(source: any = {}) {
	        return new UninstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.driversRemoved = source["driversRemoved"];
	    }
	}

}

