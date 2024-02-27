
import './style.scss';
import React, { useState, useEffect, useRef, useContext } from 'react';
import { Divider } from 'antd';

import { USER_IMAGE } from 'const/localStorageConsts';
import { httpRequest } from 'services/http';
import { ApiEndpoints } from 'const/apiEndpoints';
import { Context } from 'hooks/store';

import { ReactComponent as SearchIcon } from 'assets/images/searchIcon.svg';
import { ReactComponent as StationIcon } from 'assets/images/TotalStations.svg';
import { ReactComponent as SchemaIcon } from 'assets/images/schemaItemIcon.svg';
import { ReactComponent as MessageIcon } from 'assets/images/TotalMessages.svg';



const OverViewSearchBar = ({ open, handleClose }) => {
    const [searchforThis, setSearchforThis] = useState('');

    // These states will hold all the data fetched from server
    const [listOfStations, setListOfStations] = useState([]);
    const [listOfSchemas, setListOfSchemas] = useState([]);
    const [listOfUsers, setListOfUsers] = useState([]);
    const [listOfProducers, setListOfProducers] = useState([]);
    const [listOfMessages, setListOfMessages] = useState([]);
    const [listOfTags, setListOfTags] = useState([]);

    // This does the auto focus on input field when we open it
    const searchInputRef = useRef(null);
    useEffect(() => {
        searchInputRef.current && searchInputRef.current.focus();
    }, [open]);

    // runs only when user clicks on search icon in overview page
    useEffect(() => {
        if (open) {
            getAllData();
        }
    }, [open]);

    const getAllData = async () => {
        try {
            const resStations = await httpRequest('GET', `${ApiEndpoints.GET_STATIONS}`);
            const resSchemas = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            const resUsers = await httpRequest('GET', ApiEndpoints.GET_ALL_USERS);

            // resStations.stations.sort((a, b) => new Date(b.station.created_at) - new Date(a.station.created_at));
            const curatedApplicationUsers = resUsers.application_users.map(data => ({id: data.id, username: data.username, avatar_id: data.avatar_id, user_type: data.user_type, full_name:data.full_name, created_at:data.created_at}))
            const curatedManagementUsers = resUsers.management_users.map(data => ({id: data.id, username: data.username, avatar_id: data.avatar_id, user_type: data.user_type, full_name:data.full_name, created_at:data.created_at}))
            const curatedListOfStations = resStations.stations.map(data => ({id: data.station.id ,name: data.station.name, created_at:data.station.created_at})) 
            const curatedListOfSchemas = resSchemas.map(data => ({id: data.id, name: data.name, created_at:data.created_at})) 

            setListOfStations([...curatedListOfStations])
            setListOfSchemas([...curatedListOfSchemas])
            setListOfUsers([...curatedApplicationUsers,...curatedManagementUsers])

        } catch (err) {
            console.log(err)
            return;
        }
    };

   const filteredListOfStations = listOfStations.filter((data) => {
        return data.name.toLowerCase().includes(searchforThis.toLowerCase());
    });
    const filteredListOfProducers = listOfProducers.filter((data) => {
        return data.toLowerCase().includes(searchforThis.toLowerCase());
    });
    const filteredListOfTags = listOfTags.filter((data) => {
        return data.toLowerCase().includes(searchforThis.toLowerCase());
    });
    const filteredListOfSchemas = listOfSchemas.filter((data) => {
        return data.name.toLowerCase().includes(searchforThis.toLowerCase());
    });
    const filteredListOfUsers = listOfUsers.filter((data) => {
        return data.username.toLowerCase().includes(searchforThis.toLowerCase());
    });
    const filteredListOfMessages = listOfMessages.filter((data) => {
        return data.toLowerCase().includes(searchforThis.toLowerCase());
    });


    const getAvatarSrc = (avatarId) => {
        return (localStorage.getItem(USER_IMAGE)) || require(`assets/images/bots/avatar${avatarId}.svg`);
    };

    const listStationsDiv = (
        <div>
            <Divider />
            <div className="header">Stations</div>
            <div className="list">
                {filteredListOfStations.map((station, index) => (
                    <div key={index} className="container">
                        <div className="icon">
                            <StationIcon width={40} height={40} />
                        </div>
                        <div className="content">
                            <span className="data">{station.name}</span>
                            <div className="meta-data">{station.created_at}</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
    const listOfProducersDiv = (
        <div>
            <Divider />
            <div className="header">Producers</div>
            <div className="list">
                {listOfProducers.map((producer, index) => (
                    <div key={index} className="container">
                        <div className="icon">
                            <StationIcon />
                        </div>
                        <div className="content">
                            <span className="data">{producer}</span>
                            <div className="meta-data">time</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
    const listOfTagsDiv = (
        <div>
            <Divider />
            <div className="header">Tags</div>
            <div className="list">
                {listOfTags.map((tag, index) => (
                    <div key={index} className="container">
                        <div className="icon">
                            <StationIcon />
                        </div>
                        <div className="content">
                            <span className="data">{tag}</span>
                            <div className="meta-data">time</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
    const listOfSchemasDiv = (
        <div>
            <Divider />
            <div className="header">Schemas</div>
            <div className="list">
                {listOfSchemas.map((schema) => (
                    <div key={schema.id} className="container">
                        <div className="icon">
                            <SchemaIcon width={40} height={40} />
                        </div>
                        <div className="content">
                            <span className="data">{schema.name}</span>
                            <div className="meta-data">{schema.created_at}</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
    const listOfUsersDiv = (
        <div>
            <Divider />
            <div className="header">Users</div>
            <div className="list">
                {filteredListOfUsers.map((user) => (
                    <div key={user.id} className="container">
                        <div className="icon">
                        <img src={getAvatarSrc(user.avatar_id)} width={25} height={25} alt="avatar" />
                            {/* <StationIcon /> */}
                        </div>
                        <div className="content">
                            <span className="data">{user.username}</span>
                            <div className="meta-data">{user.created_at}</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
    const listOfMessagesDiv = (
        <div>
            <Divider />
            <div className="header">Messages</div>
            <div className="list">
                {listOfMessages.map((message, index) => (
                    <div key={index} className="container">
                        <div className="icon">
                            <MessageIcon />
                        </div>
                        <div className="content">
                            <span className="data">{message}</span>
                            <div className="meta-data">time</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );

    if (!open) return null;
    return (
        <>
            <div className="overlay-style" onClick={handleClose} />

            <div className="modal-style">
                <div className="search-bar">
                    <span>
                        {' '}
                        <SearchIcon />{' '}
                    </span>
                    <input
                        className="search-input"
                        placeholder="Search for stations, tags, producers, schemas ..."
                        ref={searchInputRef}
                        onChange={(e) => setSearchforThis(e.target.value)}
                    ></input>
                </div>
                {searchforThis.length ? (
                    <div className="all-elements">
                        {filteredListOfStations.length ? listStationsDiv : null}
                        {filteredListOfSchemas.length ? listOfSchemasDiv : null}
                        {filteredListOfProducers.length ? listOfProducersDiv : null}
                        {filteredListOfTags.length ? listOfTagsDiv : null}
                        {filteredListOfUsers.length ? listOfUsersDiv : null}
                        {filteredListOfMessages.length ? listOfMessagesDiv : null}
                    </div>
                ) : null}
            </div>
        </>
    );
};

export default OverViewSearchBar