'use strict';

import config from './config';
import { wrapError } from './utils';
import { shopClient as client, isReady } from './setup';
import uuidV4 from 'uuid/v4';

export async function getContractTypes(shopType) {
  if (!isReady()) {
    return;
  }
  try {
    return await query('contract_type_ls', { shopType });
  } catch (e) {
    throw wrapError(
      `Error getting contract types for shop type ${shopType} : ${e.message}`
      , e);
  }
}

// Alternative: getContractsOfMyShop(shopId) which returns only the contracts that are signed through my shops!
// In invoke_shop.go there should be corresponding method for returning right contracts.

export async function getContracts(username) {
  if(!isReady()) {
    return;
  }
  // omit validations of username
  try {
    if (typeof username !== 'string') {
      username = undefined;
    }
    const contracts = await query( 'shop_contract_ls', { username });
    return contracts;
  } catch (e) {
    throw wrapError(`Error getting contracts for ${username}: ${e.message}`, e);
  }
}

export async function createContract(contract) {
  if (!isReady()) {
    return;
  }
  try {
    let c = Object.assign({}, contract, { uuid: uuidV4() });
    const loginInfo = await invoke('contract_create', c);

    console.log('shopPeer.createContract c' + JSON.stringify(c));
    
    if (!loginInfo
      ^ !!(loginInfo && loginInfo.username && loginInfo.password)) {
      return Object.assign(loginInfo || {}, { uuid: c.uuid });
    } else {
      throw new Error(loginInfo);
    }
  } catch (e) {
    throw wrapError(`Error creating contract: ${e.message}`, e);
  }
}

export async function createUser(user) {
  if (!isReady()) {
    return;
  }
  try {
    const loginInfo = await invoke('user_create', user);
    if (!loginInfo ^
      !!(loginInfo && loginInfo.username && loginInfo.password)) {
      return loginInfo;
    } else {
      throw new Error(loginInfo);
    }
  } catch (e) {
    throw wrapError(`Error creating user: ${e.message}`, e);
  }
}

export async function authenticateUser(username, password) {
  if (!isReady()) {
    return;
  }
  try {
    let authenticated =
      await query('user_authenticate', { username, password });
    if (authenticated === undefined || authenticated === null) {
      throw new Error('Unknown error, invalid response!');
    }
    return authenticated;
  } catch (e) {
    throw wrapError(`Error authenticating user: ${e.message}`, e);
  }
}

export async function getUserInfo(username) {
  if (!isReady()) {
    return;
  }
  try {
    const user = await query('user_get_info', { username });
    return user;
  } catch (e) {
    throw wrapError(`Error getting user info: ${e.message}`, e);
  }
}

export function getBlocks(noOfLastBlocks) {
  return client.getBlocks(noOfLastBlocks);
}

export const on = client.on.bind(client);
export const once = client.once.bind(client);
export const addListener = client.addListener.bind(client);
export const prependListener = client.prependListener.bind(client);
export const removeListener = client.removeListener.bind(client);

function invoke(fcn, ...args) {
  return client.invoke(
    config.chaincodeId, config.chaincodeVersion, fcn, ...args);
}

function query(fcn, ...args) {
  return client.query(
    config.chaincodeId, config.chaincodeVersion, fcn, ...args);
}
