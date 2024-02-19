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

import { Paywall, useStiggContext } from '@stigg/react-sdk';
import { FiArrowUpRight, FiArrowDownLeft } from 'react-icons/fi';
import { HiQuestionMarkCircle, HiOutlineExclamationCircle } from 'react-icons/hi';
import React, { Fragment, useContext, useState } from 'react';
import { BsCheckLg } from 'react-icons/bs';
import { Link } from 'react-router-dom';

import { ReactComponent as RedirectIcon } from 'assets/images/redirectIcon.svg';
import { showMessages, useGetAllowedActions } from 'services/genericServices';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import pathDomains from 'router';
import Button from 'components/button';
import Modal from 'components/modal';
import Input from 'components/Input';
import CheckboxComponent from 'components/checkBox';
import { Context } from 'hooks/store';
import { LOCAL_STORAGE_PLAN } from 'const/localStorageConsts';
import { planType } from 'const/globalConst';

const reasons = ['Price is too high', 'Missing feature', 'Bad support', 'Performance', 'Limitations', 'Not using anymore', 'I switched to a competitor', 'Other'];

const UpgradePlans = ({ open, onClose, content, isExternal = true }) => {
    const [_, dispatch] = useContext(Context);
    const { refreshData, isInitialized } = useStiggContext();
    const [instructionsModalOpen, setInstructionsModalOpen] = useState(false);
    const [downgradeModalOpen, setDowngradeModalOpen] = useState(false);
    const [downgradeInstructions, setDowngradeInstructions] = useState({});
    const [upgradeModalOpen, setUpgradeModalOpen] = useState(false);
    const [isCheck, setIsCheck] = useState([]);
    const [downgradeLoader, setDowngradeLoader] = useState(false);
    const [downgradeReaon, setReasonDowngrade] = useState('');
    const [planSelected, setPlanSelected] = useState({});
    const [textInput, setTextInput] = useState('');
    const getAllowedActions = useGetAllowedActions();
    const success_url = window.location.href;
    const cancel_url = window.location.href;

    const handlePlanSelected = async (plan) => {
        let isDowngrade = plan.intentionType === 'DOWNGRADE_PLAN';
        setPlanSelected(plan);
        if (isDowngrade) {
            try {
                const data = await httpRequest('POST', ApiEndpoints.DOWNGRADE_CHECK, { plan: plan.plan.id });
                if (Object.keys(data?.entitlements).length > 0) {
                    setDowngradeInstructions(data?.entitlements);
                    setInstructionsModalOpen(true);
                } else {
                    setDowngradeModalOpen(true);
                }
            } catch (error) {}
        } else {
            await updatePlan(plan);
        }
    };

    const updatePlan = async (plan, withReason = false) => {
        let reason = '';
        if (withReason) {
            setDowngradeLoader(true);
            const selectedOptionsText = isCheck.join(', ');
            const enteredText = textInput.trim();
            if (selectedOptionsText && enteredText) {
                reason = `${selectedOptionsText}, ${enteredText}`;
            } else {
                reason = selectedOptionsText || enteredText;
            }
        }
        try {
            let quantity = 0;
            if (plan.billableFeatures.length > 0) {
                quantity = plan.billableFeatures[0].quantity;
            }
            const data = await httpRequest('POST', ApiEndpoints.UPGRADE_PLAN, { plan: plan.plan.id, success_url, cancel_url, reason: reason, unit_quantity: quantity });
            if (data.resp_type === 'payment') window.open(data.stripe_url, '_self');
            else {
                dispatch({ type: 'SET_ENTITLEMENTS', payload: data.entitlements });
                dispatch({ type: 'SET_PLAN_TYPE', payload: data.plan === planType.FREE });
                await refreshData();
                setTimeout(() => {
                    showMessages('success', 'Your plan has been successfully updated.');
                }, 1000);
                await getAllowedActions();
                isExternal ? onClose() : setUpgradeModalOpen(false);
                localStorage.setItem(LOCAL_STORAGE_PLAN, data.plan);
            }
        } catch (error) {
            isExternal ? onClose() : setUpgradeModalOpen(false);
        } finally {
            if (withReason) {
                setDowngradeModalOpen(false);
                setDowngradeLoader(false);
                setReasonDowngrade('');
                setIsCheck([]);
            }
        }
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        setIsCheck([...isCheck, id]);
        if (checked) {
            setReasonDowngrade({ ...downgradeReaon, id });
        }
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
        }
    };

    const handleUpdateTrigger = () => {
        if (!isInitialized) {
            showMessages('warning', 'Oh no! We are experiencing some issues with our new billing model. Please check again in a few minutes.');
        } else {
            setUpgradeModalOpen(true);
        }
    };

    return (
        <div className="upgrade-plans-container">
            {!isExternal && (
                <div
                    className="content-button-wrapper"
                    onClick={() => {
                        handleUpdateTrigger();
                    }}
                >
                    {content}
                </div>
            )}
            <Modal
                className="pricing-plans-modal"
                height="96vh"
                displayButtons={false}
                clickOutside={() => (isExternal ? onClose() : setUpgradeModalOpen(false))}
                open={isExternal ? open : upgradeModalOpen}
            >
                <Fragment>
                    <div className="paywall-header">
                        <p>Pricing & Plans</p>
                        <div className="description">
                            <label>Prices are not included traffic charges!</label>
                            <a className="a-link" href="https://memphis.dev/pricing" target="_blank">
                                Explore more
                            </a>
                            <FiArrowUpRight />
                        </div>
                    </div>
                    <Paywall onPlanSelected={(plan) => handlePlanSelected(plan)} highlightedPlanId="plan-cloud-growth-buckets" />
                    <div className="paywall-footer">
                        <label>*Coming soon</label>
                        <div className="question-info">
                            <HiQuestionMarkCircle />
                            <label>
                                Questions? We are here to help! Please reach us via the
                                <a className="a-link" href="https://memphis.dev/contact-us/?inquiry=sales">
                                    support page
                                </a>
                            </label>
                            <FiArrowUpRight />
                        </div>
                    </div>
                </Fragment>
            </Modal>
            <Modal
                className="modal-wrapper instructions-modal"
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <FiArrowDownLeft className="headerIcon" />
                        </div>
                        <p>Downgrade instructions</p>
                        <label>Your data is your most valuable asset, and we want to make sure you are aware to the limitation of your new plan.</label>
                    </div>
                }
                width="42vw"
                displayButtons={false}
                clickOutside={() => setInstructionsModalOpen(false)}
                open={instructionsModalOpen}
            >
                <Fragment>
                    <div className="instructions-redirect">
                        {downgradeInstructions['feature-integration-slack'] && (
                            <div className="redirect-section">
                                <div className="violations-list">
                                    <p className="violation-title"> Using Slack integration is violating the new plan</p>
                                    <div className="hint-line">
                                        <HiOutlineExclamationCircle />
                                        <span>Please fix the following issues before performing a downgrade</span>
                                    </div>
                                </div>
                            </div>
                        )}
                        {downgradeInstructions['feature-integration-discord'] && (
                            <div className="redirect-section">
                                <div className="violations-list">
                                    <p className="violation-title"> Using Discord integration is violating the new plan</p>
                                    <div className="hint-line">
                                        <HiOutlineExclamationCircle />
                                        <span>Please fix the following issues before performing a downgrade</span>
                                    </div>
                                </div>
                            </div>
                        )}
                        {downgradeInstructions['feature-management-users'] && (
                            <div className="redirect-section">
                                <p className="violation-title">Too many management users ({downgradeInstructions['feature-management-users']['usage']})</p>
                                <div className="hint-line">
                                    <HiOutlineExclamationCircle />
                                    <span>
                                        The plan you are about to downgrade to allows {downgradeInstructions['feature-management-users']['limits']} management users
                                    </span>
                                </div>
                                <Link to={pathDomains.users}>
                                    <div
                                        className="flex-line"
                                        onClick={() => {
                                            setInstructionsModalOpen(false);
                                            isExternal ? onClose() : setUpgradeModalOpen(false);
                                        }}
                                    >
                                        <span>Go to User Management</span>
                                        <RedirectIcon />
                                    </div>
                                </Link>
                            </div>
                        )}
                        {(downgradeInstructions['feature-partitions-per-station'] ||
                            downgradeInstructions['feature-storage-retention'] ||
                            downgradeInstructions['feature-storage-tiering'] ||
                            downgradeInstructions['feature-stations-limitation']) && (
                            <div className="redirect-section">
                                <p className="violation-title">Some stations are violating the new plan </p>
                                <div className="hint-line">
                                    <HiOutlineExclamationCircle />
                                    <span>Please fix the following issues before performing a downgrade</span>
                                </div>
                                <Link to={pathDomains.stations}>
                                    <div
                                        className="flex-line"
                                        onClick={() => {
                                            setInstructionsModalOpen(false);
                                            isExternal ? onClose() : setUpgradeModalOpen(false);
                                        }}
                                    >
                                        <span>Go to Station list</span>
                                        <RedirectIcon />
                                    </div>
                                </Link>
                                {downgradeInstructions['feature-storage-retention'] && (
                                    <div className="violations-list">
                                        <p className="violation-title">Some retention policy is violating the new plan</p>
                                        <div className="hint-line">
                                            <HiOutlineExclamationCircle />
                                            <span>
                                                The plan you are about to downgrade to allows {downgradeInstructions['feature-storage-retention']?.limits} retention days
                                                per station
                                            </span>
                                        </div>
                                    </div>
                                )}
                                {downgradeInstructions['feature-storage-tiering'] && (
                                    <div className="violations-list">
                                        <p className="violation-title">Storage tiering is activated</p>
                                        <div className="hint-line">
                                            <HiOutlineExclamationCircle />
                                            <span>The plan you are about to downgrade to does not allow to use storage tiering</span>
                                        </div>
                                    </div>
                                )}
                                {downgradeInstructions['feature-partitions-per-station'] && (
                                    <div className="violations-list">
                                        <p className="violation-title">Some partitions policy is violating the new plan</p>
                                        <div className="hint-line">
                                            <HiOutlineExclamationCircle />
                                            <span>
                                                The plan you are about to downgrade to allows {downgradeInstructions['feature-partitions-per-station']?.limits} partitons
                                                per station
                                            </span>
                                        </div>
                                    </div>
                                )}
                                {downgradeInstructions['feature-stations-limitation'] && (
                                    <div className="violations-list">
                                        <p className="violation-title">You have more stations than the allowed amount of stations in the new plan</p>
                                        <div className="hint-line">
                                            <HiOutlineExclamationCircle />
                                            <span>
                                                The plan you are about to downgrade to allows {downgradeInstructions['feature-stations-limitation']?.limits} stations
                                            </span>
                                        </div>
                                    </div>
                                )}
                                {downgradeInstructions['feature-apply-functions'] && (
                                    <div className="violations-list">
                                        <p className="violation-title">Some stations have attached functions</p>
                                        <div className="hint-line">
                                            <HiOutlineExclamationCircle />
                                            <span>The plan you are about to downgrade does not allow to use attached functions feature</span>
                                        </div>
                                    </div>
                                )}
                                {downgradeInstructions['feature-dls-consumption-linkage'] && (
                                    <div className="violations-list">
                                        <p className="violation-title">Some stations are utilizing the DLS linkage feature</p>
                                        <div className="hint-line">
                                            <HiOutlineExclamationCircle />
                                            <span>The plan you are about to downgrade to does not allow to use DLS linkage feature</span>
                                        </div>
                                    </div>
                                )}
                                </div>
                            )}
                    </div>
                    <div className="instructions-button">
                        <Button
                            className="modal-btn"
                            width="120px"
                            height="32px"
                            placeholder="Done"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'purple'}
                            fontSize="12px"
                            fontWeight="600"
                            onClick={() => {
                                setInstructionsModalOpen(false);
                            }}
                        />
                    </div>
                </Fragment>
            </Modal>
            <Modal
                className="modal-wrapper downgrade-modal"
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <BsCheckLg className="headerIcon" />
                        </div>
                        <p>Downgrade Successful</p>
                        <label>We understand your decision and would love to know how to improve ourselves</label>
                    </div>
                }
                width="30vw"
                displayButtons={false}
                open={downgradeModalOpen}
                clickOutside={() => setDowngradeModalOpen(false)}
            >
                <Fragment>
                    <div className="downgrade-reasons">
                        {reasons.map((reason) => (
                            <CheckboxComponent key={reason} checked={isCheck?.includes(reason)} id={reason} onChange={handleCheckedClick} checkName={reason} />
                        ))}
                    </div>
                    {isCheck?.includes('Other') && (
                        <div className="downgrade-box">
                            <p>Tell us why you decided to downgrade</p>
                            <Input
                                placeholder="Tell us why you decided to downgrade"
                                type="textArea"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                numberOfRows={4}
                                fontSize="14px"
                                value={textInput}
                                onBlur={(e) => setTextInput(e.target.value)}
                                onChange={(e) => setTextInput(e.target.value)}
                            />
                        </div>
                    )}
                    <div className="instructions-button">
                        <Button
                            className="modal-btn"
                            width="10vw"
                            height="32px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            border="gray-light"
                            backgroundColorType={'white'}
                            fontSize="12px"
                            fontWeight="600"
                            onClick={() => {
                                setDowngradeModalOpen(false);
                            }}
                        />
                        <Button
                            className="modal-btn"
                            width="10vw"
                            height="32px"
                            placeholder="Downgrade"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'purple'}
                            fontSize="12px"
                            fontWeight="600"
                            disabled={isCheck?.length === 0}
                            onClick={() => {
                                updatePlan(planSelected, true);
                            }}
                            isLoading={downgradeLoader}
                        />
                    </div>
                </Fragment>
            </Modal>
        </div>
    );
};
export default UpgradePlans;
