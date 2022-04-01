/**
 * Flamenco manager
 * Render Farm manager API
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 *
 */

import ApiClient from '../ApiClient';

/**
 * The ManagerConfiguration model module.
 * @module model/ManagerConfiguration
 * @version 0.0.0
 */
class ManagerConfiguration {
    /**
     * Constructs a new <code>ManagerConfiguration</code>.
     * @alias module:model/ManagerConfiguration
     * @param storageLocation {String} Directory used for job file storage.
     * @param shamanEnabled {Boolean} Whether the Shaman file transfer API is available.
     */
    constructor(storageLocation, shamanEnabled) { 
        
        ManagerConfiguration.initialize(this, storageLocation, shamanEnabled);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, storageLocation, shamanEnabled) { 
        obj['storageLocation'] = storageLocation;
        obj['shamanEnabled'] = shamanEnabled;
    }

    /**
     * Constructs a <code>ManagerConfiguration</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/ManagerConfiguration} obj Optional instance to populate.
     * @return {module:model/ManagerConfiguration} The populated <code>ManagerConfiguration</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new ManagerConfiguration();

            if (data.hasOwnProperty('storageLocation')) {
                obj['storageLocation'] = ApiClient.convertToType(data['storageLocation'], 'String');
            }
            if (data.hasOwnProperty('shamanEnabled')) {
                obj['shamanEnabled'] = ApiClient.convertToType(data['shamanEnabled'], 'Boolean');
            }
        }
        return obj;
    }


}

/**
 * Directory used for job file storage.
 * @member {String} storageLocation
 */
ManagerConfiguration.prototype['storageLocation'] = undefined;

/**
 * Whether the Shaman file transfer API is available.
 * @member {Boolean} shamanEnabled
 */
ManagerConfiguration.prototype['shamanEnabled'] = undefined;






export default ManagerConfiguration;

