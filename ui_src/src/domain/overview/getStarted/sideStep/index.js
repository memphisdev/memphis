// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import React from 'react';
import './style.scss';
import GetStartedIcon from '../../../../assets/images/getStartedIcon.svg';
import AppUserIcon from '../../../../assets/images/usersIconActive.svg';
import EmptyStation from '../../../../assets/images/emptyStation.svg';
import DataProduced from '../../../../assets/images/dataProduced.svg';
import ConsumeDataImg from '../../../../assets/images/stationsIconActive.svg';
import FullStation from '../../../../assets/images/fullStation.svg';
import FinishFlag from '../../../../assets/images/finishFlag.svg';
import GrayAppUserIcon from '../../../../assets/images/grayAppUserIcon.svg';
import GrayProduceDataImg from '../../../../assets/images/grayProduceDataImg.svg';
import GrayConsumeDataImg from '../../../../assets/images/grayConsumeDataImg.svg';
import GrayfinishStep from '../../../../assets/images/grayFinish.svg';
import CompletedStep from '../../../../assets/images/completedStep.svg';

const SideStep = (props) => {
    const { stepNumber, stepName, currentStep, completedSteps, stepsDescription } = props;

    const getDocLink = () => {
        switch (stepNumber) {
            case 1:
                return 'https://docs.memphis.dev/memphis-new/dashboard-ui/stations';
            case 2:
                return 'https://docs.memphis.dev/memphis-new/dashboard-ui/users';
            case 3:
                return 'https://docs.memphis.dev/memphis-new/memphis/concepts/producer';
            case 4:
                return 'https://docs.memphis.dev/memphis-new/memphis/concepts/consumer';
            default:
                return;
        }
    };
    const getIcon = () => {
        switch (stepNumber) {
            case 1:
                return <img className="sidebar-image" src={GetStartedIcon} alt="getStartedIcon" />;
            case 2:
                return completedSteps + 1 >= stepNumber ? (
                    <img className="sidebar-image" src={AppUserIcon} alt="getStartedIcon" />
                ) : (
                    <img className="sidebar-image" src={GrayAppUserIcon} alt="getStartedIcon" />
                );
            case 3:
                if (completedSteps + 1 > stepNumber) return <img className="sidebar-image" src={DataProduced} alt="getStartedIcon" />;
                else if (completedSteps + 1 === stepNumber) return <img className="sidebar-image" src={EmptyStation} alt="getStartedIcon" />;
                else return <img className="sidebar-image" src={GrayProduceDataImg} alt="getStartedIcon" />;
            case 4:
                if (completedSteps + 1 > stepNumber) return <img className="sidebar-image" src={ConsumeDataImg} alt="getStartedIcon" />;
                else if (completedSteps + 1 === stepNumber) return <img className="sidebar-image" src={FullStation} alt="getStartedIcon" />;
                else return <img className="sidebar-image" src={GrayConsumeDataImg} alt="getStartedIcon" />;
            case 5:
                return completedSteps + 1 >= stepNumber ? (
                    <img className="sidebar-image" src={FinishFlag} alt="getStartedIcon" />
                ) : (
                    <img className="sidebar-image" src={GrayfinishStep} alt="getStartedIcon" />
                );
            default:
                return;
        }
    };
    return (
        <div
            className={completedSteps + 1 >= stepNumber ? 'side-step-container cursor-allowed' : 'side-step-container'}
            onClick={() => completedSteps + 1 >= stepNumber && props.onSideBarClick(stepNumber)}
        >
            <div className="side-step-header">
                {getIcon()}
                <div className="step-name-completed">
                    <p className={currentStep === stepNumber ? 'step-name curr-step-name' : 'step-name'}>{stepName}</p>
                    {completedSteps >= stepNumber && stepNumber !== 5 && <img className="completed" src={CompletedStep} alt="completed" />}
                </div>
            </div>
            <div className={completedSteps >= stepNumber ? 'side-step-body border-completed' : stepNumber !== 5 ? 'side-step-body border' : 'side-step-body'}>
                {stepNumber !== 5 && (
                    <p className={currentStep === stepNumber ? 'step-description curr-step-name' : 'step-description'}>
                        {stepsDescription}
                        {'. '}
                        <a href={getDocLink()} target="_blank" rel="noopener noreferrer">
                            Learn more
                        </a>
                    </p>
                )}
            </div>
        </div>
    );
};
export default SideStep;
