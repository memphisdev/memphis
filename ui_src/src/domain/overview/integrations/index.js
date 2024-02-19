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

import React, { useEffect, useState, useRef, useContext } from 'react';
import { useHistory } from 'react-router-dom';

import { Context } from 'hooks/store';
import debeziumIcon from 'assets/images/debeziumIcon.svg';
import slackLogo from 'assets/images/slackLogo.svg';
import discordIntegrationIcon from 'assets/images/discordIntegrationIcon.svg';
import s3Logo from 'assets/images/s3Logo.svg';
import pathDomains from 'router';
import Modal from 'components/modal';
import SlackIntegration from '../../administration/integrations/components/slackIntegration';
import DiscordIntegration from '../../administration/integrations/components/discordIntegration';
import S3Integration from '../../administration/integrations/components/s3Integration';
import DebeziumIntegration from '../../administration/integrations/components/debeziumIntegration';
import { httpRequest } from 'services/http';
import { ApiEndpoints } from 'const/apiEndpoints';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import ErrorRoundedIcon from '@material-ui/icons/ErrorRounded';
import LockFeature from 'components/lockFeature';
import {entitlementChecker} from "utils/plan";

const Integrations = () => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrations, setIntegrations] = useState([
        { name: 'Slack', logo: slackLogo, value: {} },
        { name: 'Discord', logo: discordIntegrationIcon, value: {} },
        { name: 'S3', logo: s3Logo, value: {} },
        { name: 'Debezium', logo: debeziumIcon, value: {} }
    ]);
    const history = useHistory();
    const ref = useRef();
    const storageTiringLimits = !entitlementChecker(state, 'feature-storage-tiering');

    useEffect(() => {
        getallIntegration();
    }, []);

    const getallIntegration = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_INTEGRATION);
            dispatch({ type: 'SET_INTEGRATIONS', payload: data || [] });
        } catch (err) {
            return;
        }
    };

    const updateIntegrationValue = (value, index) => {
        let integrationsCopy = [...integrations];
        integrationsCopy[index].value = value;
        setIntegrations(integrationsCopy);
    };
    useEffect(() => {
        if (state.integrationsList?.length > 0) {
            integrations?.forEach((integration, index) => {
                const value = checkIfUsed(integration?.name);
                if (value) {
                    updateIntegrationValue(value, index);
                }
            });
        }
    }, [state?.integrationsList]);

    const checkIfUsed = (integrationName) => {
        let index = state.integrationsList?.findIndex((integration) => integration.name === integrationName?.toLowerCase());
        if (index !== -1) {
            return state.integrationsList[index];
        } else return;
    };

    const isValidIndication = (indicator) => {
        return indicator ? <CheckCircleIcon className="connected" /> : <ErrorRoundedIcon className="broken" />;
    };

    const modalContent = () => {
        switch (ref.current) {
            case 'Slack':
                return (
                    <SlackIntegration
                        close={(value) => {
                            modalFlip(false);
                            updateIntegrationValue(value, 0);
                        }}
                        value={integrations[0]?.value}
                    />
                );
            case 'Discord':
                return (
                    <DiscordIntegration
                        close={(value) => {
                            modalFlip(false);
                            updateIntegrationValue(value, 0);
                        }}
                        value={integrations[0]?.value}
                    />
                );
            case 'S3':
                return (
                    <S3Integration
                        close={(value) => {
                            modalFlip(false);
                            updateIntegrationValue(value, 1);
                        }}
                        value={integrations[1]?.value}
                    />
                );
            case 'Debezium':
                return (
                    <DebeziumIntegration
                        close={(value) => {
                            modalFlip(false);
                            updateIntegrationValue(value, 2);
                        }}
                        value={integrations[2]?.value}
                    />
                );
            default:
                break;
        }
    };
    return (
        <div className="overview-components-wrapper">
            <div className="overview-integrations-container">
                <div className="overview-components-header integrations-header">
                    <p>Integrations</p>
                    <label className="link-to-page" onClick={() => history.push(`${pathDomains.administration}/integrations`)}>
                        Explore more Integrations
                    </label>
                </div>
                <div className="integrations-list">
                    {integrations?.map((integration, index) => {
                        return (
                            <div
                                className="integration-item"
                                key={index}
                                onClick={() => {
                                    if (storageTiringLimits && integration.name === 'S3') {
                                        return;
                                    } else {
                                        ref.current = integration.name;
                                        modalFlip(true);
                                    }
                                }}
                            >
                                {storageTiringLimits && integration.name === 'S3' ? (
                                    <LockFeature />
                                ) : (
                                    integrations[index]?.value &&
                                    Object.keys(integrations[index]?.value).length > 0 &&
                                    isValidIndication(integrations[index]?.value?.is_valid)
                                )}
                                <img className="img-icon" src={integration.logo} alt={integration.name} />
                                <label className="integration-name">{integration.name}</label>
                            </div>
                        );
                    })}
                </div>
            </div>
            <Modal className="integration-modal" height="96vh" width="720px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                {modalContent()}
            </Modal>
        </div>
    );
};

export default Integrations;
