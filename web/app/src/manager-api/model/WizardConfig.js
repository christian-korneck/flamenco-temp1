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
import BlenderPathCheckResult from './BlenderPathCheckResult';

/**
 * The WizardConfig model module.
 * @module model/WizardConfig
 * @version 0.0.0
 */
class WizardConfig {
    /**
     * Constructs a new <code>WizardConfig</code>.
     * Configuration obtained from the First-Time Wizard.
     * @alias module:model/WizardConfig
     * @param storageLocation {String} Directory used for job file storage.
     * @param blenderExecutable {module:model/BlenderPathCheckResult} 
     */
    constructor(storageLocation, blenderExecutable) { 
        
        WizardConfig.initialize(this, storageLocation, blenderExecutable);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, storageLocation, blenderExecutable) { 
        obj['storageLocation'] = storageLocation;
        obj['blenderExecutable'] = blenderExecutable;
    }

    /**
     * Constructs a <code>WizardConfig</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/WizardConfig} obj Optional instance to populate.
     * @return {module:model/WizardConfig} The populated <code>WizardConfig</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new WizardConfig();

            if (data.hasOwnProperty('storageLocation')) {
                obj['storageLocation'] = ApiClient.convertToType(data['storageLocation'], 'String');
            }
            if (data.hasOwnProperty('blenderExecutable')) {
                obj['blenderExecutable'] = BlenderPathCheckResult.constructFromObject(data['blenderExecutable']);
            }
        }
        return obj;
    }


}

/**
 * Directory used for job file storage.
 * @member {String} storageLocation
 */
WizardConfig.prototype['storageLocation'] = undefined;

/**
 * @member {module:model/BlenderPathCheckResult} blenderExecutable
 */
WizardConfig.prototype['blenderExecutable'] = undefined;






export default WizardConfig;

