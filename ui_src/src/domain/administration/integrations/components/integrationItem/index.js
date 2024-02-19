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

import React, { useState, useEffect, useContext, useRef } from 'react';
import { useHistory } from 'react-router-dom';
import { ReactComponent as IntegratedIcon } from 'assets/images/integrated.svg';
import { ReactComponent as IntegrationFailedIcon } from 'assets/images/integrationFailed.svg';
import { ReactComponent as MemphisVerifiedIcon } from 'assets/images/memphisFunctionIcon.svg';
import { capitalizeFirst } from 'services/valueConvertor';
import { Context } from 'hooks/store';
import SlackIntegration from '../slackIntegration';
import DiscordIntegration from '../discordIntegration';
import S3Integration from '../s3Integration';
import Tag from 'components/tag';
import DataDogIntegration from '../dataDogIntegration';
import GrafanaIntegration from '../grafanaIntegration';
import ElasticIntegration from '../elasticIntegration';
import DebeziumIntegration from '../debeziumIntegration';
import GitHubIntegration from '../gitHubIntegration';
import ZapierIntegration from '../zapierIntegration';
import Drawer from "components/drawer";

const IntegrationItem = ({ value, lockFeature, isOpen }) => {
    const [state] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrateValue, setIntegrateValue] = useState({});

    const ref = useRef();
    ref.current = integrateValue;
    const history = useHistory();

    useEffect(() => {
        modalFlip(isOpen);
    }, [isOpen]);

    useEffect(() => {
        if (state.integrationsList?.length > 0) {
            checkIfUsed();
        }
    }, [state?.integrationsList]);

    const checkIfUsed = () => {
        let index = state.integrationsList?.findIndex((integration) => capitalizeFirst(integration.name) === value?.name);
        setIntegrateValue(state.integrationsList[index]);
    };

    const modalContent = () => {
        switch (value?.name) {
            case 'Slack':
                return (
                    <SlackIntegration
                        close={(data) => {
                            modalFlip(false);
                            setIntegrateValue(data);
                        }}
                        value={ref.current}
                    />
                );
            case 'Discord':
                return (
                    <DiscordIntegration
                        close={(data) => {
                            modalFlip(false);
                            setIntegrateValue(data);
                        }}
                        value={ref.current}
                    />
                );
            case 'Github':
                return (
                    <GitHubIntegration
                        close={(data) => {
                            setIntegrateValue(data);
                            data !== ref.current &&
                                history.push({
                                    pathname: '/functions',
                                    integrated: true
                                });
                        }}
                        value={ref.current}
                    />
                );
            case 'S3':
                return (
                    <S3Integration
                        close={(data) => {
                            modalFlip(false);
                            setIntegrateValue(data);
                        }}
                        value={ref.current}
                        lockFeature={lockFeature}
                    />
                );
            case 'Datadog':
                return (
                    <DataDogIntegration
                        close={() => {
                            modalFlip(false);
                        }}
                    />
                );
            case 'Grafana':
                return (
                    <GrafanaIntegration
                        close={() => {
                            modalFlip(false);
                        }}
                    />
                );
            case 'Debezium and Postgres':
                return (
                    <DebeziumIntegration
                        close={() => {
                            modalFlip(false);
                        }}
                    />
                );

            case `Elasticsearch observability`:
                return (
                    <ElasticIntegration
                        close={() => {
                            modalFlip(false);
                        }}
                    />
                );

            case `Zapier`:
                return (
                    <ZapierIntegration
                        close={() => {
                            modalFlip(false);
                        }}
                    />
                );

            default:
                break;
        }
    };

    return (
        <>
            <integ-item is="3xd" onClick={() => (value?.comingSoon ? null : modalFlip(true))}>
                {value?.banner}
                {integrateValue && Object.keys(integrateValue)?.length !== 0 && integrateValue?.is_valid && (
                    <div className="integrate-icon">
                        <IntegratedIcon />
                        <p>Integrated</p>
                    </div>
                )}
                {integrateValue && Object.keys(integrateValue)?.length !== 0 && !integrateValue?.is_valid && (
                    <div className="broken-integration-icon">
                        <IntegrationFailedIcon />
                        <p>Integration Failed</p>
                    </div>
                )}
                <div className="integration-name">
                    {value?.icon}
                    <div className="details">
                        <p>{value?.name}</p>
                        <span className="by">
                            <MemphisVerifiedIcon />
                            <label className="memphis">{value?.by}</label>
                        </span>
                    </div>
                </div>
                <p className="integration-description">{value?.description} </p>
                <div className="category">
                    <Tag tag={value?.category} />
                </div>
            </integ-item>
            <Drawer
                placement="right"
                onClose={() => {
                    if (value?.name === 'Github') {
                        if (state.integrationsList?.findIndex((integration) => capitalizeFirst(integration.name) === 'Github') !== -1) {
                            history.push({
                                pathname: '/functions',
                                integrated: true
                            });
                        }
                    }
                    modalFlip(false);
                }}
                destroyOnClose={true}
                className="integration-modal"
                width="720px"
                open={modalIsOpen}
                closeIcon={false}
                headerStyle={{ display: 'none' }}
                bodyStyle={{ padding: '0px' }}
            >
                {modalContent()}
            </Drawer>
        </>
    );
};

export default IntegrationItem;
