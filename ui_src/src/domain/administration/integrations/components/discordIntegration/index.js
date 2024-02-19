// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useState, useContext, useEffect } from 'react';
import { Form } from 'antd';

import { ReactComponent as PoisionAlertIcon } from 'assets/images/poisionAlertIcon.svg';
import { ReactComponent as DisconAlertIcon } from 'assets/images/disconAlertIcon.svg';
import { ReactComponent as SchemaAlertIcon } from 'assets/images/schemaAlertIcon.svg';
import { ReactComponent as PurpleQuestionMark } from 'assets/images/purpleQuestionMark.svg';
import { INTEGRATION_LIST, getTabList } from 'const/integrationList';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import Switcher from 'components/switcher';
import Button from 'components/button';
import { Context } from 'hooks/store';
import Input from 'components/Input';
import CustomTabs from 'components/Tabs';
import { URL } from 'config';
import Loader from 'components/loader';
import CloudMoadl from 'components/cloudModal';
import { showMessages } from 'services/genericServices';
import { isCloud } from 'services/valueConvertor';
import IntegrationDetails from '../integrationItem/integrationDetails';
import IntegrationLogs from '../integrationItem/integrationLogs';
import { FaArrowCircleUp } from 'react-icons/fa';

const urlSplit = URL.split('/', 3);

const DiscordIntegration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
    const discordConfiguration = INTEGRATION_LIST['Discord'];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        name: 'discord',
        ui_url: `${urlSplit[0]}//${urlSplit[2]}`,
        keys: {
            webhook_url: value?.keys?.webhook_url || '',
        },
        properties: {
            poison_message_alert: value?.properties ? (value?.properties?.poison_message_alert ? true : false) : true,
            schema_validation_fail_alert: value?.properties ? (value?.properties?.schema_validation_fail_alert ? true : false) : true,
            disconnection_events_alert: value?.properties ? (value?.properties?.disconnection_events_alert ? true : false) : true
        }
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [tabValue, setTabValue] = useState('Configuration');
    const [cloudModalOpen, setCloudModalOpen] = useState(false);
    const tabs = getTabList('Discord');

    useEffect(() => {
        const images = [];
        images.push(INTEGRATION_LIST['Discord'].banner.props.src);
        images.push(INTEGRATION_LIST['Discord'].insideBanner.props.src);
        const promises = [];

        images.forEach((imageUrl) => {
            const image = new Image();
            promises.push(
                new Promise((resolve) => {
                    image.onload = resolve;
                })
            );
            image.src = imageUrl;
        });

        Promise.all(promises).then(() => {
            setImagesLoaded(true);
        });
    }, []);

    const updateKeysState = (field, value) => {
        let updatedValue = { ...formFields.keys };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };
    const updatePropertiesState = (field, value) => {
        let updatedValue = { ...formFields.properties };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, properties: updatedValue }));
    };

    const handleSubmit = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            setLoadingSubmit(true);
            if (isValue) {
                if (values.webhook_url === 'https://discord.com/api/webhooks/****') {
                    updateIntegration(false);
                } else {
                    updateIntegration();
                }
            } else {
                createIntegration();
            }
        }
    };

    const closeModal = (data, disconnect = false) => {
        setTimeout(() => {
            disconnect ? setLoadingDisconnect(false) : setLoadingSubmit(false);
        }, 1000);
        close(data);
        showMessages('success', disconnect ? 'The integration was successfully disconnected' : 'The integration connected successfully');
    };

    const updateIntegration = async (withWebhook = true) => {
        let newFormFields = { ...formFields };
        if (!withWebhook) {
            let updatedKeys = { ...formFields.keys };
            updatedKeys['webhook_url'] = '';
            newFormFields = { ...newFormFields, keys: updatedKeys };
        }
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, { ...newFormFields });
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const createIntegration = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, { ...formFields });
            dispatch({ type: 'ADD_INTEGRATION', payload: data });
            closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };
    const disconnect = async () => {
        setLoadingDisconnect(true);
        try {
            await httpRequest('DELETE', ApiEndpoints.DISCONNECT_INTEGRATION, {
                name: formFields.name
            });
            dispatch({ type: 'REMOVE_INTEGRATION', payload: formFields.name });

            closeModal({}, true);
        } catch (err) {
            setLoadingDisconnect(false);
        }
    };

    return (
        <dynamic-integration is="3xd" className="integration-modal-container">
            {!imagesLoaded && (
                <div className="loader-integration-box">
                    <Loader />
                </div>
            )}
            {imagesLoaded && (
                <>
                    {discordConfiguration?.insideBanner}
                    <div className="integrate-header">
                        {discordConfiguration.header}
                        <div className={!isValue ? 'action-buttons flex-end' : 'action-buttons'}>
                            {isValue && (
                                <Button
                                    width="100px"
                                    height="35px"
                                    placeholder="Disconnect"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="red"
                                    border="none"
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    isLoading={loadingDisconnect}
                                    onClick={() => disconnect()}
                                />
                            )}
                            <PurpleQuestionMark
                                className="info-icon"
                                alt="Integration info"
                                onClick={() => window.open('https://docs.memphis.dev/memphis/dashboard-ui/integrations/notifications', '_blank')}
                            />
                        </div>
                    </div>
                    <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs} />

                    <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                        {tabValue === 'Details' && <IntegrationDetails integrateDesc={discordConfiguration.integrateDesc} />}
                        {tabValue === 'Logs' && <IntegrationLogs integrationName={'discord'} />}
                        {tabValue === 'Configuration' && (
                            <div className="integration-body">
                                <IntegrationDetails integrateDesc={discordConfiguration.integrateDesc} />

                                <div className="api-details">
                                    <p className="title">API details</p>
                                    <div className="api-key">
                                        <p>Webhook URL</p>
                                        <span className="desc">Copy and paste your discord 'Webhook URL' here</span>
                                        <Form.Item
                                            name="webhook_url"
                                            rules={[
                                                {
                                                    required: true,
                                                    message: 'Please insert webhook URL.'
                                                }
                                            ]}
                                            initialValue={formFields?.keys?.webhook_url}
                                        >
                                            <Input
                                                placeholder="https://discord.com/api/webhooks/****"
                                                type="text"
                                                radiusType="semi-round"
                                                colorType="black"
                                                backgroundColorType="purple"
                                                borderColorType="none"
                                                height="40px"
                                                fontSize="12px"
                                                onBlur={(e) => updateKeysState('webhook_url', e.target.value)}
                                                onChange={(e) => updateKeysState('webhook_url', e.target.value)}
                                                value={formFields?.keys?.webhook_url}
                                            />
                                        </Form.Item>
                                    </div>
                                    
                                    <div className="notification-option">
                                        <p>Notify me when:</p>
                                        <span className="desc">Memphis will send only the selected triggers</span>
                                        <>
                                            <div className="option-wrapper">
                                                <div className="option-name">
                                                    <PoisionAlertIcon />
                                                    <div className="name-des">
                                                        <p>New unacked message</p>
                                                        <span>
                                                            Messages that cause a consumer group to repeatedly require a delivery (possibly due to a consumer failure)
                                                            such that the message is never processed completely and acknowledged
                                                        </span>
                                                    </div>
                                                </div>
                                                <Form.Item name="poison_message_alert">
                                                    <Switcher
                                                        onChange={() => updatePropertiesState('poison_message_alert', !formFields.properties.poison_message_alert)}
                                                        checked={formFields.properties?.poison_message_alert}
                                                    />
                                                </Form.Item>
                                            </div>
                                            <div className="option-wrapper">
                                                <div className="option-name">
                                                    <SchemaAlertIcon />
                                                    <div className="name-des">
                                                        <p>Schema validation failure</p>
                                                        <span>Triggered once a client fails in schema validation</span>
                                                    </div>
                                                </div>
                                                <Form.Item name="schema_validation_fail_alert">
                                                    <Switcher
                                                        onChange={() =>
                                                            updatePropertiesState('schema_validation_fail_alert', !formFields.properties.schema_validation_fail_alert)
                                                        }
                                                        checked={formFields.properties?.schema_validation_fail_alert}
                                                    />
                                                </Form.Item>
                                            </div>
                                            <div className="option-wrapper">
                                                <div className="option-name">
                                                    <DisconAlertIcon />
                                                    <div className="name-des">
                                                        <p>Disconnected clients</p>
                                                        <span>Triggered once a producer/consumer get disconnected</span>
                                                    </div>
                                                </div>
                                                <Form.Item name="schema_validation_fail_alert">
                                                    <Switcher
                                                        onChange={() =>
                                                            updatePropertiesState('disconnection_events_alert', !formFields.properties.disconnection_events_alert)
                                                        }
                                                        checked={formFields.properties?.disconnection_events_alert}
                                                    />
                                                </Form.Item>
                                            </div>
                                        </>
                                    </div>
                                </div>
                            </div>
                        )}
                        <Form.Item className="button-container">
                            <div className="button-wrapper">
                                <Button
                                    width="500px"
                                    height="45px"
                                    placeholder={
                                        isValue ? (
                                            'Update'
                                        ) : isCloud() && !state?.allowedActions?.can_connect_discord ? (
                                            <span className="upgrade">
                                                Connect <FaArrowCircleUp className="lock-feature-icon" />
                                            </span>
                                        ) : (
                                            'Connect'
                                        )
                                    }
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="14px"
                                    fontFamily="InterSemiBold"
                                    isLoading={loadingSubmit}
                                    disabled={isValue && !creationForm.isFieldsTouched()}
                                    onClick={() => (!isCloud() || state?.allowedActions?.can_connect_discord ? handleSubmit() : setCloudModalOpen(true))}
                                />
                            </div>
                        </Form.Item>
                    </Form>
                    <CloudMoadl type={'upgrade'} open={cloudModalOpen} handleClose={() => setCloudModalOpen(false)} />
                </>
            )}
        </dynamic-integration>
    );
};

export default DiscordIntegration;
