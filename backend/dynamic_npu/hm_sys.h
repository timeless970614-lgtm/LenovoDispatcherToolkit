// SPDX-License-Identifer: GPL-2.0
/*
 * Copyright(c) 2025 Houmo AI Inc.
 * Author: liang.huang<liang.huang@houmo.ai>
 */

#ifndef _HM_SYS_H_
#define _HM_SYS_H_

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

#ifndef _SIZE_T
#include <stddef.h>
#endif

#include "hm_api.h"

/**
 * @brief Maximum supported length (in bytes) of a model
 *        name string.
 *
 * This macro defines the recommended size of buffers used
 * to store Houmo model names.
 */
#define HM_SYS_DEVICE_NAME_LEN 16

/**
 * @brief Maximum supported length (in bytes) of a device
 *        serial number string.
 *
 * This macro defines the recommended size of buffers used
 * to store Houmo device serial number.
 */
#define HM_SYS_DEVICE_SN_LEN 32

/**
 * @brief Maximum number of Houmo devices that can be connected
 *        to the host.
 *
 * This macro defines the upper limit for the number of logical
 * devices supported by the system.
 */
#define HM_MAX_DEVICES 32

/**
 * @brief Structure that holds information about Houmo devices.
 *
 */
struct hm_device_info {
	/// The number of Houmo devices currently connected to the host.
	uint32_t num_devices;
	/// Array of logical IDs for each Houmo device.
	uint32_t device_ids[HM_MAX_DEVICES];
};

/**
 * @brief Retrieves information about all Houmo devices
 *        connected to the host.
 *
 * @param info[out] Pointer to an hm_device_info structure that
 *        holds the device information.
 * @return Returns the number of Houmo devices on success; returns
 *         0 on failure.
 */
HM_HAL_API
uint32_t hm_sys_get_device_info(struct hm_device_info *info);

/**
 * @brief Checks if the specified Houmo device index is valid.
 *
 * This function verifies that the given device index exists in
 * the list of detected Houmo devices.
 *
 * @param dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @return Returns 0 if the device index is valid; returns -1 if
 *         the device index is invalid.
 * @see #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_check_device_index(int dev_index);

/**
 * @brief Retrieves the vendor ID of a Houmo device with the
 *        specified logical ID.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       The vendor ID can be retrieved only if the eFuse has been
 *       programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @return Returns the vendor ID of the specified Houmo device on
 *         success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_vendor_id(int dev_index);

/**
 * @brief Retrieves the serial number of a Houmo device with the
 *        specified logical ID.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       The serial number can be retrieved only if the eFuse has
 *       been programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] sn A pointer to a buffer for storing the serial number.
 * @param[in] len The maximum size of the output buffer in bytes.
 *             The maximum value is HM_SYS_DEVICE_SN_LEN.
 *             It is recommended to set this value to
 *             HM_SYS_DEVICE_SN_LEN. If the buffer size is smaller
 *             than the serial number length, the function returns
 *             an error.
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_device_sn(int dev_index, char sn[], int len);

/**
 * @brief Retrieves the model name of a Houmo device with the
 *        specified logical ID.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       The model name can be retrieved only if the eFuse has
 *       been programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] name A pointer to a buffer for storing the model name.
 * @param[in] len The maximum size of the output buffer in bytes.
 *             The maximum value is HM_SYS_DEVICE_NAME_LEN.
 *             It is recommended to set this value to
 *             HM_SYS_DEVICE_NAME_LEN. If the buffer size is smaller
 *             than the model name length, the function returns
 *             an error.
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_device_name(int dev_index, char name[], int len);

/**
 * @brief Retrieves the computing power (TOPS) of the specified
 *        Houmo device.
 *
 * @note This function requires valid data in the Houmo device
 *       eFuse. The computing power can be retrieved only if the
 *       eFuse has been programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @return Returns the computing power on success; returns -1
 *         on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_ipu_utili_rate,
 *      #hm_sys_get_ipu_frequency, #hm_sys_get_board_power,
 *      #hm_sys_get_core_count, #hm_sys_get_ipu_voltage,
 *      #hm_sys_get_temperature, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_computing_power(int dev_index);

/**
 * @brief Retrieves the total number of IPU cores of the specified
 *        Houmo device.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       The number of IPU cores can be retrieved only if the eFuse
 *       has been programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @return Returns the total number of IPU cores on success; returns
 *         -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_core_count(int dev_index);

/**
 * @brief Retrieves the real-time average utilization rate of all
 *        IPU cores of the specified Houmo device.
 *
 * The return value is shown as a floating-point value in the
 * range 0.0 to 1.0.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       IPU core utilization rate can be retrieved only if the
 *       eFuse has been programmed.
 *
 * @param [in] dev_index The logical ID of the Houmo device. Valid IDs
 *             can be get from `struct hm_device_info::device_ids`.
 * @return Returns the average utilization rate of all IPU cores
 *         on success; returns a negative value on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_computing_power,
 *      #hm_sys_get_ipu_frequency, #hm_sys_get_board_power,
 *      #hm_sys_get_core_count, #hm_sys_get_ipu_voltage,
 *      #hm_sys_get_temperature, #hm_sys_check_device_index,
 *      #hm_sys_get_ipu_core_utili_rate
 */
HM_HAL_API
float hm_sys_get_ipu_utili_rate(int dev_index);

/**
 * @brief Retrieves the real-time utilization rate of the specified
 *        IPU core on the specified Houmo device.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       IPU core utilization rate can be retrieved only if the
 *       eFuse has been programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[in] core_id The ID of the IPU core for which the utilization
 *            rate is retrieved.
 * @return Returns the real-time utilization rate of the specified
 *         IPU cores on success; returns a negative value on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_computing_power,
 *      #hm_sys_get_ipu_voltage, #hm_sys_get_board_power,
 *      #hm_sys_get_ipu_utili_rate, #hm_sys_get_core_count,
 *      #hm_sys_get_temperature, #hm_sys_check_device_index
 */
HM_HAL_API
float hm_sys_get_ipu_core_utili_rate(int dev_index, uint32_t core_id);

/**
 * @brief Retrieves the real-time average voltage (mV) of all IPU
 *        cores of the specified Houmo device.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] voltage Pointer to the voltage value in millivolts (mV).
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_computing_power,
 *      #hm_sys_get_ipu_frequency, #hm_sys_get_board_power,
 *      #hm_sys_get_core_count, #hm_sys_get_ipu_utili_rate,
 *      #hm_sys_get_temperature, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_ipu_voltage(int dev_index, float *voltage);

/**
 * @brief Retrieves the real-time average frequency (Hz) of all
 *        IPU cores of the specified Houmo device.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] frequency Pointer to the frequency value
 *             in Hertz (Hz).
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_computing_power,
 *      #hm_sys_get_ipu_voltage, #hm_sys_get_board_power,
 *      #hm_sys_get_ipu_utili_rate, #hm_sys_get_core_count,
 *      #hm_sys_get_temperature, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_ipu_frequency(int dev_index, uint64_t *frequency);

/**
 * @struct hm_mem_info
 * @brief Structure that holds DDR memory information of a Houmo device.
 */
struct hm_mem_info {
	/// Total DDR memory of the Houmo device in MB.
	uint32_t mem_total;
	/// DDR memory consumed by the Houmo device in MB.
	uint32_t mem_used;
	/// DDR memory available for the Houmo device in MB.
	uint32_t mem_avail;
};

/**
 * @brief Retrieves the DDR memory information of the specified
 *        Houmo device.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] mem_info Pointer to an hm_mem_info struct that holds
 *             the DDR memory information.
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_ddr_size,
 *      #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_mem_info(int dev_index, struct hm_mem_info *mem_info);

/**
 * @brief Retrieves the real-time temperature (°C) of the specified
 *        Houmo device.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] temp Pointer to the temperature value in Celsius (°C).
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_computing_power,
 *      #hm_sys_get_ipu_voltage, #hm_sys_get_ipu_frequency,
 *      #hm_sys_get_ipu_utili_rate, #hm_sys_get_core_count,
 *      #hm_sys_get_board_power, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_temperature(int dev_index, float *temp);

/**
 * @brief Retrieves the build time of the Houmo Dadao SDK as a
 *        string in the format ``YYYY-MM-DD HH:MM:SS``.
 *
 * @param[out] buildtime A pointer to a buffer for storing the
 *             build time string.
 * @param[in] len The maximum size of the output buffer in bytes.
 *             The minimum supported value is 32 bytes.
 * @return Returns 0 on success; returns -1 on failure.
 */
HM_HAL_API
int hm_sys_get_buildtime(char *buildtime, size_t len);

/**
 * @brief Retrieves the version of the Houmo Dadao SDK.
 *
 * @param[out] version A pointer to a buffer for storing the
 *             version string.
 * @param[in] len The maximum size of the output buffer in bytes.
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_driver_version, #hm_sys_get_device_version
 */
HM_HAL_API
int hm_sys_get_version(char *version, size_t len);

/**
 * @brief Retrieves the driver version string.
 *
 * @param[out] version A pointer to a buffer for storing the driver
 *             version string.
 * @param[in] len The maximum size of the output buffer in bytes.
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_version, #hm_sys_get_device_version
 */
HM_HAL_API
int hm_sys_get_driver_version(char *version, size_t len);

/**
 * @brief Retrieves the firmware version of the specified Houmo device.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] version A pointer to a buffer for storing the firmware
 *             version string.
 * @param[in] len The maximum size of the output buffer in bytes.
 * @return Returns 0 on success; returns -1 on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_version,
 *      #hm_sys_get_driver_version, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_device_version(int dev_index, char *version, size_t len);

/**
 * @brief Retrieves the total DDR memory size (in bytes) of the
 *        specified Houmo device.
 *
 * @note This function requires valid data in the Houmo device eFuse.
 *       The total DDR memory size can be retrieved only if the eFuse has
 *       been programmed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] ddr_size Pointer to the total DDR memory size in bytes.
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_mem_info,
 *      #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_ddr_size(int dev_index, uint64_t *ddr_size);

/**
 * @brief Retrieves the PCIe BDF (Bus, Device, Function) of the
 *        specified Houmo device.
 *
 * The BDF string uniquely identifies the PCIe device within the system.
 *
 * Linux Format: ``domain:bus:device.function``
 *
 * Components:
 *
 * - ``domain``: PCIe domain number (Linux only).
 * - ``bus``: PCIe bus number.
 * - ``device``: Device number on the specified bus.
 * - ``function``: Specific function or port number of the device.
 *
 * @note This function is supported only on Linux and Android systems.
 * It is not supported on Windows.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] bdf A pointer to a buffer for storing the PCIe BDF string.
 * @param[in] len The maximum size of the output buffer in bytes.
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_get_device_info, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_bdf(int dev_index, char *bdf, size_t len);

/**
 * @brief Retrieves the real-time PCIe link bandwidth of the
 *        specified Houmo device.
 *
 * This function retrieves the current data transfer capability
 * of the PCIe interface connecting the device to the host system.
 *
 * The bandwidth is returned in the format ``<link_rate>-<lane_count>``,
 * for example, ``5.0 GT/s-4lane``:
 *
 * - ``link_rate``: Data transfer rate per lane in gigatransfers per
 *                  second (GT/s).
 * - ``lane_count``: Number of active PCIe lanes.
 *
 * @note This function is supported only on Linux and Android systems.
 *       It is not supported on Windows.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] bandwidth A pointer to a buffer for storing the PCIe link
 *             bandwidth string.
 * @param[in] len The maximum size of the output buffer in bytes.
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_get_device_info, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_bandwidth(int dev_index, char *bandwidth, size_t len);

/**
 * @brief Retrieves the real-time power consumption (W) of the
 *        specified Houmo device board.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] power A pointer to a float to store the power consumption
 *             in Watts (W).
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_get_device_info, #hm_sys_get_computing_power,
 *      #hm_sys_get_core_count, #hm_sys_get_ipu_voltage,
 *      #hm_sys_get_ipu_frequency, #hm_sys_get_temperature,
 *      #hm_sys_get_ipu_utili_rate, #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_board_power(int dev_index, float *power);

/**
 * @brief Enumerated type indicating the DVFS (Dynamic Voltage and
 *        Frequency Scaling) mode used for the sepcified Houmo
 *        logical device.
 *
 * DVFS controls the operating frequency of IPU cores to balance
 * performance and power consumption.
 */
enum hm_dvfs_mode {
	/// IPU cores always run at maximum frequency of 1400 MHz.
	/// This is the default mode.
	HM_DVFS_PERFORMANCE = 0,
	/// IPU core frequency is dynamically adjusted based on real-time
	/// utilization within the default range of 1400 MHz to 200 MHz.
	HM_DVFS_ONDEMAND,
	/// Sentinel value representing the total number of DVFS modes.
	/// This value is not a valid mode. It is used for iteration
	/// or bounds checking.
	HM_DVFS_MODE_MAX,
};

/**
 * @brief Retrieves the current DVFS mode of the specified Houmo device.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] mode A pointer to an #hm_dvfs_mode enum to store the current
 *             DVFS mode.
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_set_dvfs_mode, #hm_sys_get_device_info,
 *      #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_get_dvfs_mode(int dev_index, enum hm_dvfs_mode *mode);

/**
 * @brief Sets the DVFS mode for the specified M50 logical
 *        device of a Houmo product.
 *
 * **Note**
 *
 * - After a power cycle, the system returns to the default
 *   #HM_DVFS_PERFORMANCE mode.
 * - The DVFS mode is configured per M50 logical device. Each M50 chip
 *   is treated as an M50 logical device. If a Houmo product contains
 *   more than one M50 chip, this function sets the DVFS mode only
 *   for the M50 logical device specifed by the ``dev_index`` parameter.
 *
 * @param[in] dev_index The logical ID of the Houmo M50 chip. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[in] mode The DVFS mode to set. Valid values are
 *            #HM_DVFS_PERFORMANCE and #HM_DVFS_ONDEMAND.
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_get_dvfs_mode, #hm_sys_get_device_info,
 *      #hm_sys_check_device_index
 */
HM_HAL_API
int hm_sys_set_dvfs_mode(int dev_index, enum hm_dvfs_mode mode);

/**
 * @brief Retrieves the CTC (Chip-to-Chip) PHY identifiers associated
 *        with the given Houmo device ID.
 *
 * In configurations with multiple Houmo chips, Houmo chips
 * interconnected via CTC form a logical group. Each group
 * represents an independent CTC interconnection domain.
 * Only chips belonging to the same group can establish physical
 * CTC connections.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[out] group_id Pointer to the group ID.
 * @param[out] chip_id Pointer to the chip ID.
 * @return Returns 0 on success; returns a negative value on failure.
 * @see #hm_sys_check_device_index
 *
 */
HM_HAL_API
int hm_sys_get_ctc_phy_id(int dev_index, int *group_id, int *chip_id);

/**
 * @brief Enumerated type indicating the status of a firmware upgrade
 *        operation.
 */
typedef enum {
	/// Upgrade operation succeeded.
	HM_FW_UPGRADE_SUCCESS = 0,

	/// Invalid input parameter. Ensure all input values are valid.
	HM_FW_UPGRADE_INVALID_PARAM,

	/// Memory allocation failed during upgrade. Ensure enough
	/// memory is available for the firmware upgrade.
	HM_FW_UPGRADE_MEM_ALLOC_FAILED,

	/// Failed to read/stat/seek firmware image file. Verify the permission
	/// of the firmware image.
	HM_FW_UPGRADE_FILE_OPER_FAILED,

	/// Invalid firmware image header. Verify the firmware header format
	/// and integrity.
	HM_FW_UPGRADE_INVALID_HEADER,

	/// Invalid firmware magic number. Check firmware compatibility and
	/// integrity.
	HM_FW_UPGRADE_INVALID_MAGIC,

	/// Invalid metadata partition index. Check firmware integrity.
	HM_FW_UPGRADE_INVALID_META,

	/// Invalid partition image size in firmware. Check firmware integrity.
	HM_FW_UPGRADE_INVALID_IMAGE_SIZE,

	/// Invalid firmware image CRC checksum. The firmware may be corrupted
	/// or tampered; verify the firmware integrity.
	HM_FW_UPGRADE_INVALID_IMAGE_CRC,

	/// Device operation failed (e.g., read memory error).
	/// Check device connection, power supply, and device status.
	HM_FW_UPGRADE_DEVICE_OPER_FAILED,

	/// Failed to program the firmware image into the device Flash
	/// memory. Verify device connection and available Flash space.
	HM_FW_UPGRADE_FLASH_PROGRAM_FAILED,

	/// Failed to create the firmware upgrade thread.
	HM_FW_UPGRADE_THREAD_ERROR,

	/// A firmware upgrade operation is in progress.
	HM_FW_UPGRADE_BUSY,

	/// Maximum result code (invalid, used for boundary checking only).
	HM_FW_UPGRADE_RESULT_MAX,
} hm_fw_upgrade_result_t;

/**
 * @brief Firmware upgrade context.
 *
 * @note This is an opaque type. Applications must not access or
 *       modify its internal members directly.
 *
 */
typedef struct hm_fw_upgrade_ctx hm_fw_upgrade_ctx_t;

/**
 * @brief Initializes a firmware upgrade context.
 *
 * **Note**
 *
 * The context is allocated on the heap and must be released by calling
 * hm_fw_upgrade_deinit() when no longer needed.
 *
 * @param[in] dev_index The logical ID of the Houmo device. Valid IDs
 *            can be get from `struct hm_device_info::device_ids`.
 * @param[in] fw_img Path to the firmware image ``firmware.img``.
 * @return Returns a pointer to the initialized firmware upgrade
 *         context on success; returns NULL on failure.
 * @see #hm_fw_upgrade_deinit, #hm_fw_upgrade_execute,
 *      #hm_fw_upgrade_start, #hm_fw_upgrade_get_result,
 *      #hm_fw_upgrade_get_progress
 */
HM_HAL_API
hm_fw_upgrade_ctx_t *hm_fw_upgrade_init(int dev_index, const char *device_path);

/**
 * @brief Destroys the firmware upgrade context and releases the upgrade resources.
 *
 * @param[in] ctx The firmware upgrade context to destroy.
 * @return Returns #HM_FW_UPGRADE_SUCCESS if the context is successfully
 *         destroyed; returns a #hm_fw_upgrade_result_t error code on failure.
 * @see #hm_fw_upgrade_init, #hm_fw_upgrade_execute,
 *      #hm_fw_upgrade_start, #hm_fw_upgrade_get_result,
 *      #hm_fw_upgrade_get_progress
 */
HM_HAL_API
hm_fw_upgrade_result_t hm_fw_upgrade_deinit(hm_fw_upgrade_ctx_t *ctx);

/**
 * @brief Executes a firmware upgrade synchronously.
 *
 * This function performs the firmware upgrade in the calling thread
 * and blocks until the upgrade process completes.
 *
 * This API is used in the following scenarios:
 *
 * - Single-device synchronous upgrade.
 * - User-managed thread upgrade, where the caller creates a thread and
 *   invokes this function within that thread.
 *
 * @note You must call hm_fw_upgrade_init() to initialize the firmware upgrade
 *       context before calliing this API, and call hm_fw_upgrade_deinit() after
 *       the upgrade process completes to destroy the upgrade context and release
 *       resources.
 *
 * @param[in] ctx The firmware upgrade context.
 * @return Returns #HM_FW_UPGRADE_SUCCESS indicates successful completion;
 *         returns a #hm_fw_upgrade_result_t error code on failure.
 * @see #hm_fw_upgrade_init, #hm_fw_upgrade_deinit,
 *      #hm_fw_upgrade_start, #hm_fw_upgrade_get_result,
 *      #hm_fw_upgrade_get_progress
 */
HM_HAL_API
hm_fw_upgrade_result_t hm_fw_upgrade_execute(hm_fw_upgrade_ctx_t *ctx);

/**
 * @brief Starts a firmware upgrade in asynchronous mode.
 *
 * This function starts the firmware upgrade process and returns immediately.
 * The upgrade is executed by an internal worker thread, which internally
 * invokes ::hm_fw_upgrade_execute.
 *
 * This API is used in the following scenarios:
 *
 * - Single-device asynchronous upgrade
 * - Multi-device parallel upgrade, where one upgrade context is created
 *   per device and this function is called for each context
 *
 * @note You must call hm_fw_upgrade_init() to initialize the firmware upgrade
 *       context before calliing this API, and call hm_fw_upgrade_deinit() after
 *       the upgrade process completes to destroy the upgrade context and release
 *       resources.
 * @param[in] ctx  The firmware upgrade context.
 * @return Returns a #hm_fw_upgrade_result_t value indicating if the upgrade
 *         process was sucessfully started.
 * @see #hm_fw_upgrade_init, #hm_fw_upgrade_deinit,
 *      #hm_fw_upgrade_execute, #hm_fw_upgrade_get_result,
 *      #hm_fw_upgrade_get_progress
 */
HM_HAL_API
hm_fw_upgrade_result_t hm_fw_upgrade_start(hm_fw_upgrade_ctx_t *ctx);

/**
 * @brief Retrieve the firmware upgrade operation result.
 *
 * This function returns the result code of the firmware upgrade
 * operation associated with the specified context.
 *
 * @param[in] ctx  The firmware upgrade context.
 * @return Returns a #hm_fw_upgrade_result_t value indicating the current
 *         status of the firmware upgrade operation.
 * @see #hm_fw_upgrade_init, #hm_fw_upgrade_deinit,
 *      #hm_fw_upgrade_execute, #hm_fw_upgrade_start,
 *      #hm_fw_upgrade_get_progress
 */
HM_HAL_API
hm_fw_upgrade_result_t hm_fw_upgrade_get_result(hm_fw_upgrade_ctx_t *ctx);

/**
 * @brief Retrieves the current firmware upgrade progress.
 *
 * This function retrieves the current progress of an ongoing firmware
 * upgrade operation in a thread-safe manner. Progress is reported
 * on a scale from 0 to 10000, corresponding to 0.00% to 100.00%.
 *
 * @param[in] ctx  The firmware upgrade context.
 * @param[out] progress The current upgrade progress value
 * @return Returns a #hm_fw_upgrade_result_t value indicating on
 *         success or failure.
 * @see #hm_fw_upgrade_init, #hm_fw_upgrade_deinit,
 *      #hm_fw_upgrade_execute, #hm_fw_upgrade_start,
 *      #hm_fw_upgrade_get_result
 */
HM_HAL_API
hm_fw_upgrade_result_t hm_fw_upgrade_get_progress(hm_fw_upgrade_ctx_t *ctx,
	int *progress);

/**
 * @brief Retrieves the firmware version from a firmware image.
 *
 * The firmware version consists of major, minor and patch components.
 * For instance, if major is 2, minor is 0, and patch is 0, this API
 * the string "v2.0.0".
 *
 * @param[in] fw_path Path to the firmware image ``firmware.img``.
 * @param[out] version Pointer to an output buffer used to store the
 *             firmware version string. On success, the string is
 *             null-terminated.
 * @param[in] len The maximum size of the output buffer in bytes.
 * @return Returns 0 on success; returns a negative value on failure.
 */
HM_HAL_API
int hm_fw_get_version(const char *fw_path, char version[], size_t len);
#ifdef __cplusplus
}
#endif
#endif /* _HM_SYS_H_ */
