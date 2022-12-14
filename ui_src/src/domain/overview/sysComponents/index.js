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

import './style.scss';

import React, { useContext } from 'react';
import { Divider } from '@material-ui/core';

import HealthyBadge from '../../../components/healthyBadge';
import { Context } from '../../../hooks/store';
import { PieChart, Pie } from 'recharts';
import OverflowTip from '../../../components/tooltip/overflowtip';
import { Add } from '@material-ui/icons';
import { Popover } from 'antd';

const remainingPorstPopInnerStyle = { padding: '10px', borderRadius: '12px', border: '1px solid #f0f0f0' };

const SysComponents = () => {
    const ports = ['90000', '5555', '1234', '232323'];
    const [state, dispatch] = useContext(Context);
    const getData = (comp) => {
        let data = [];
        if (comp?.actual_pods > 0) {
            for (let i = 0; i < comp?.actual_pods; i++) data.push({ name: `actual${i}`, value: 1, fill: '#6557FF' });
        }
        if (comp?.desired_pods > comp?.actual_pods) {
            for (let i = 0; i < comp?.desired_pods - comp?.actual_pods; i++) data.push({ name: `desired${i}`, value: 1, fill: '#EBEAED' });
        }
        return data;
    };
    return (
        <div className="overview-wrapper sys-components-container">
            <span className="overview-components-header">System components</span>
            <div className="sys-components sys-components-header">
                <p>Component</p>
                <p>Containers</p>
                <p>Ports</p>
                <p>Status</p>
            </div>
            {!state?.monitor_data?.system_components && <Divider />}
            <div className="component-list">
                {state?.monitor_data?.system_components &&
                    state?.monitor_data?.system_components?.map((comp, i) => {
                        return (
                            <div key={`${comp.podName}${i}`}>
                                <Divider />

                                <div className="sys-components">
                                    <OverflowTip text={comp.component}>
                                        <p>{comp.component}</p>
                                    </OverflowTip>
                                    <div className="pods-container">
                                        <PieChart height={35} width={35}>
                                            <Pie dataKey="value" data={getData(comp)} startAngle={-270}></Pie>
                                        </PieChart>
                                        <p>
                                            {comp.actual_pods}/{comp.desired_pods}
                                        </p>
                                    </div>

                                    <div className="pods-container">
                                        <p>{ports[0]}</p>
                                        {ports.length > 1 && (
                                            <Popover
                                                overlayInnerStyle={remainingPorstPopInnerStyle}
                                                placement="bottomLeft"
                                                content={ports?.slice(1)?.map((port) => {
                                                    return <p className="port-popover">{port}</p>;
                                                })}
                                            >
                                                <div className="plus-ports">
                                                    <Add className="add" />
                                                    <p>{ports.length - 1}</p>
                                                </div>
                                            </Popover>
                                        )}
                                    </div>
                                    <HealthyBadge status={comp.actual_pods / comp.desired_pods} />
                                </div>
                            </div>
                        );
                    })}
            </div>
        </div>
    );
};

export default SysComponents;
