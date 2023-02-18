// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useState } from 'react';

import { compareObjects } from '../../../services/valueConvertor';
import ConfImg1 from '../../../assets/images/confImg1.svg';
import ConfImg2 from '../../../assets/images/confImg2.svg';
import ConfImg3 from '../../../assets/images/confImg3.svg';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import Button from '../../../components/button';
import SliderRow from './components/sliderRow';
import InputRow from './components/inputRow';
import TieredInputRow from './components/tieredInputRow';
import { message } from 'antd';
import {
    LOCAL_STORAGE_BROKER_HOST,
    LOCAL_STORAGE_ENV,
    LOCAL_STORAGE_REST_GW_HOST,
    LOCAL_STORAGE_UI_HOST,
    LOCAL_STORAGE_TIERED_STORAGE_TIME
} from '../../../const/localStorageConsts';

function ClusterConfiguration() {
    const tsTimeOptions = ['Seconds', 'Minutes'];
    const [formFields, setFormFields] = useState({});
    const [oldValues, setOldValues] = useState({});
    const [isChanged, setIsChanged] = useState(false);
    const [tsTimeType, setTsTimeType] = useState(tsTimeOptions[0]);

    useEffect(() => {
        getConfigurationValue();
    }, []);

    const getConfigurationValue = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_CONFIGURATION);
            setOldValues(data);
            setFormFields(data);
        } catch (err) {
            return;
        }
    };

    const updateConfiguration = async () => {
        try {
            let updatedValue = { ...formFields };
            if (tsTimeType === tsTimeOptions[1]) {
                let val = updatedValue['tiered_storage_time_sec'];
                val = val * 60;
                updatedValue['tiered_storage_time_sec'] = val;
                setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
                setTsTimeType(tsTimeOptions[0]);
            }
            const data = await httpRequest('PUT', ApiEndpoints.EDIT_CLUSTER_CONFIGURATION, { ...updatedValue });
            localStorage.setItem(LOCAL_STORAGE_BROKER_HOST, formFields.broker_host);
            localStorage.setItem(LOCAL_STORAGE_REST_GW_HOST, formFields.rest_gw_host);
            localStorage.setItem(LOCAL_STORAGE_UI_HOST, formFields.ui_host);
            localStorage.setItem(LOCAL_STORAGE_TIERED_STORAGE_TIME, formFields.tiered_storage_time_sec);

            setIsChanged(false);
            setOldValues(data);
            message.success({
                key: 'memphisSuccessMessage',
                content: 'Successfully updated',
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisSuccessMessage')
            });
        } catch (err) {
            return;
        }
    };

    const handleChange = (field, value, err, type) => {
        if (err !== '') {
            setIsChanged(false);
        } else {
            let updatedValue = { ...formFields };
            if (field === 'tiered_storage_time_sec') {
                if (type === tsTimeOptions[1]) {
                    updatedValue[field] = value * 60;
                } else {
                    updatedValue[field] = value;
                }
            }
            setIsChanged(!compareObjects(updatedValue, oldValues));
            updatedValue[field] = value;
            setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
        }
    };
    const discardChanges = () => {
        setIsChanged(false);
        setFormFields((formFields) => ({ ...formFields, ...oldValues }));
    };

    return (
        <div className="configuration-container">
            <div className="header">
                <p className="main-header">Cluster configuration</p>
                <p className="sub-header">In this section, you can tune 'Memphis' internal configuration to suit your requirements</p>
            </div>
            <div className="configuration-body">
                <SliderRow
                    title="DEAD LETTER MESSAGES RETENTION IN HOURS"
                    desc="Amount of hours to retain dead letter messages in a DLS"
                    value={formFields?.pm_retention}
                    img={ConfImg2}
                    min={1}
                    max={30}
                    unit={'h'}
                    onChanges={(e) => handleChange('pm_retention', e)}
                />
                <SliderRow
                    title="LOGS RETENTION IN DAYS"
                    desc="Amount of days to retain system logs"
                    img={ConfImg1}
                    value={formFields?.logs_retention}
                    min={1}
                    max={100}
                    unit={'d'}
                    onChanges={(e) => handleChange('logs_retention', e)}
                />
                <TieredInputRow
                    title="TIERED STORAGE UPLOAD INTERVAL"
                    desc="Interval of uploading messages to TS after retention end"
                    img={ConfImg1}
                    value={formFields?.tiered_storage_time_sec}
                    tsType={tsTimeType}
                    onChanges={(e, t, err) => {
                        setTsTimeType(t);
                        handleChange('tiered_storage_time_sec', e, err, t);
                    }}
                />
                {localStorage.getItem(LOCAL_STORAGE_ENV) !== 'docker' && !process.env.REACT_APP_SANDBOX_ENV && (
                    <>
                        <InputRow
                            title="BROKER HOST"
                            desc="BROKER HOSTBROKER HOSTBROKER HOSTBROKER HOST"
                            img={ConfImg3}
                            value={formFields?.broker_host}
                            onChanges={(e) => handleChange('broker_host', e.target.value)}
                            placeholder={localStorage.getItem(LOCAL_STORAGE_BROKER_HOST) === undefined ? localStorage.getItem(LOCAL_STORAGE_BROKER_HOST) : ''}
                        />
                        <InputRow
                            title="UI HOST"
                            desc="UI HOSTUI HOSTUI HOSTUI HOSTUI HOSTUI HOST"
                            img={ConfImg3}
                            value={formFields?.ui_host}
                            onChanges={(e) => handleChange('ui_host', e.target.value)}
                            placeholder={localStorage.getItem(LOCAL_STORAGE_UI_HOST) === undefined ? localStorage.getItem(LOCAL_STORAGE_UI_HOST) : ''}
                        />
                        <InputRow
                            title="REST HOST"
                            desc="REST HOSTREST HOSTREST HOSTREST HOSTREST HOST"
                            img={ConfImg3}
                            value={formFields?.rest_gw_host}
                            onChanges={(e) => handleChange('rest_gw_host', e.target.value)}
                            placeholder={localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST) === undefined ? localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST) : ''}
                        />
                    </>
                )}
            </div>
            <div className="configuration-footer">
                <div className="btn-container">
                    <Button
                        className="modal-btn"
                        width="100px"
                        height="34px"
                        placeholder="Discard"
                        colorType="gray-dark"
                        radiusType="circle"
                        backgroundColorType="none"
                        border="gray"
                        boxShadowsType="gray"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        disabled={!isChanged}
                        onClick={() => discardChanges()}
                    />
                    <Button
                        className="modal-btn"
                        width="100px"
                        height="34px"
                        placeholder="Apply"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        aria-haspopup="true"
                        disabled={!isChanged}
                        onClick={() => updateConfiguration()}
                    />
                </div>
            </div>
        </div>
    );
}

export default ClusterConfiguration;
