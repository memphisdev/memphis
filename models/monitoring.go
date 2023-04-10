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
package models

import "time"

type SysComponent struct {
	Name    string    `json:"name"`
	CPU     CompStats `json:"cpu"`
	Memory  CompStats `json:"memory"`
	Storage CompStats `json:"storage"`
	Healthy bool      `json:"healthy"`
	Status  string    `json:"status"`
}

type CompStats struct {
	Total      float64 `json:"total"`
	Current    float64 `json:"current"`
	Percentage int     `json:"percentage"`
}

type Components struct {
	UnhealthyComponents []SysComponent `json:"unhealthy_components"`
	DangerousComponents []SysComponent `json:"dangerous_components"`
	RiskyComponents     []SysComponent `json:"risky_components"`
	HealthyComponents   []SysComponent `json:"healthy_components"`
}
type SystemComponents struct {
	Name        string     `json:"name"`
	Components  Components `json:"components"`
	Status      string     `json:"status"`
	Ports       []int      `json:"ports"`
	DesiredPods int        `json:"desired_pods"`
	ActualPods  int        `json:"actual_pods"`
	Hosts       []string   `json:"hosts"`
}

type SystemComponentsStatus struct {
	Status         string `json:"status"`
	HealthyCount   int    `json:"healthy_count"`
	UnhealthyCount int    `json:"unhealthy_count"`
	DangerousCount int    `json:"dangerous_count"`
	RiskyCount     int    `json:"risky_count"`
}

type MainOverviewData struct {
	TotalStations     int                        `json:"total_stations"`
	TotalMessages     uint64                     `json:"total_messages"`
	TotalDlsMessages  uint64                     `json:"total_dls_messages"`
	SystemComponents  []SystemComponents         `json:"system_components"`
	Stations          []ExtendedStation          `json:"stations"`
	K8sEnv            bool                       `json:"k8s_env"`
	BrokersThroughput []BrokerThroughputResponse `json:"brokers_throughput"`
	MetricsEnabled    bool                       `json:"metrics_enabled"`
}

type GetStationOverviewDataSchema struct {
	StationName string `form:"station_name" json:"station_name"  binding:"required"`
}

type SystemLogsRequest struct {
	LogType  string `form:"log_type" json:"log_type"  binding:"required"`
	StartIdx int    `form:"start_index" json:"start_index"  binding:"required"`
}

type Log struct {
	MessageSeq int       `json:"message_seq"`
	Type       string    `json:"type"`
	Source     string    `json:"source"`
	Data       string    `json:"data"`
	TimeSent   time.Time `json:"created_at"`
}

type SystemLogsResponse struct {
	Logs []Log `json:"logs"`
}

type RestGwMonitoringResponse struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Storage float64 `json:"storage"`
}

type BrokerThroughput struct {
	Name  string `json:"name"`
	Read  int64  `json:"read"`
	Write int64  `json:"write"`
}

type BrokerThroughputResponse struct {
	Name  string                    `json:"name"`
	Read  []ThroughputReadResponse  `json:"read"`
	Write []ThroughputWriteResponse `json:"write"`
}

type ThroughputReadResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Read      int64     `json:"read"`
}

type ThroughputWriteResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Write     int64     `json:"write"`
}

type Throughput struct {
	Bytes       int64 `json:"bytes"`
	BytesPerSec int64 `json:"bytes_per_sec"`
}

type GetProducersByStationResponse struct {
	ConnectedProducers    []ExtendedProducer `json:"connected_producers"`
	DisconnectedProducers []ExtendedProducer `json:"disconnected_producers"`
	DeletedProducers      []ExtendedProducer `json:"deleted_producers"`
}

type MessagesSampleResponse struct {
	TotalMessages int              `json:"total_messages"`
	AvgMsgSize    int64            `json:"avg_msg_size"`
	Messages      []MessageDetails `json:"messages"`
}

type GetDlsMsgsByStationLightResponse struct {
	PoisonMessages       []LightDlsMessageResponse `json:"poison_messages"`
	SchemaFailedMessages []LightDlsMessageResponse `json:"schema_fail_messages"`
	TotalDlsAmount       int                       `json:"total_dls_amount"`
}

type GetCgsByStationResponse struct {
	ConnectedCgs    []Cg `json:"connected_cgs"`
	DisconnectedCgs []Cg `json:"disconnected_cgs"`
	DeletedCgs      []Cg `json:"deleted_cgs"`
}

type GetLeaderAndFollowersResponse struct {
	Leader    string   `json:"leader"`
	Followers []string `json:"followers"`
}

type GetStationOverviewDataResponse struct {
	// StationName        StationName                      `json:"station_name"`
	Station            Station                          `json:"station"`
	Producers          GetProducersByStationResponse    `json:"producers"`
	AuditLogs          []AuditLog                       `json:"audit_log"`
	MessagesSample     MessagesSampleResponse           `json:"message_sample"`
	DlsMessages        GetDlsMsgsByStationLightResponse `json:"dls_messages"`
	Cgs                GetCgsByStationResponse          `json:"cgs"`
	Tags               []CreateTag                      `json:"tags"`
	LeaderAndFollowers GetLeaderAndFollowersResponse    `json:"leader_and_followers"`
}
