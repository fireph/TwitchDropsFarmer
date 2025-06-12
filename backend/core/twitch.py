from __future__ import annotations

import json
import asyncio
import logging
from datetime import datetime, timezone, timedelta
from typing import Any, Dict, List, Optional, Union
from pathlib import Path

import aiohttp
from yarl import URL

from core.constants import (
    ClientType, 
    GQL_OPERATIONS, 
    COOKIES_PATH, 
    WATCH_INTERVAL,
    JsonType
)

logger = logging.getLogger("TwitchDrops")

class LoginException(Exception):
    pass

class RequestException(Exception):
    pass

class GQLException(Exception):
    pass

class AuthManager:
    def __init__(self, twitch_client):
        self._twitch = twitch_client
        self._logged_in = asyncio.Event()
        self.user_id: Optional[str] = None
        self.access_token: Optional[str] = None
        self.device_id: Optional[str] = None
        self.session_id: Optional[str] = None
        self.client_version: Optional[str] = None

    def _hasattrs(self, *attrs: str) -> bool:
        return all(hasattr(self, attr) for attr in attrs)

    def _delattrs(self, *attrs: str) -> None:
        for attr in attrs:
            if hasattr(self, attr):
                delattr(self, attr)

    def clear(self) -> None:
        self._delattrs(
            "user_id",
            "device_id", 
            "session_id",
            "access_token",
            "client_version",
        )
        self._logged_in.clear()

    async def wait_until_login(self) -> None:
        await self._logged_in.wait()

    @property
    def is_logged_in(self) -> bool:
        return self._logged_in.is_set() and self.access_token is not None


class Twitch:
    def __init__(self):
        self._client_type = ClientType.WEB
        self._session: Optional[aiohttp.ClientSession] = None
        self.auth = AuthManager(self)
        
    async def get_session(self) -> aiohttp.ClientSession:
        if self._session is None or self._session.closed:
            connector = aiohttp.TCPConnector(limit=16, ttl_dns_cache=60)
            timeout = aiohttp.ClientTimeout(total=10)
            
            # Load cookies if they exist
            jar = aiohttp.CookieJar()
            if COOKIES_PATH.exists():
                jar.load(COOKIES_PATH)
                
            self._session = aiohttp.ClientSession(
                connector=connector,
                timeout=timeout,
                cookie_jar=jar,
                headers={"User-Agent": self._client_type.USER_AGENT}
            )
        return self._session

    async def close_session(self):
        if self._session and not self._session.closed:
            await self._session.close()

    async def request(self, method: str, url: str, **kwargs) -> aiohttp.ClientResponse:
        session = await self.get_session()
        
        # Add authorization header if logged in
        if self.auth.access_token:
            headers = kwargs.get("headers", {})
            headers["Authorization"] = f"OAuth {self.auth.access_token}"
            kwargs["headers"] = headers
            
        return await session.request(method, url, **kwargs)

    async def gql_request(self, operations: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Make GraphQL request using exact same format as TwitchDropsMiner"""
        if not self.auth.is_logged_in:
            raise GQLException("Not logged in")
            
        session = await self.get_session()
        
        # Prepare payload - single operation or list
        if len(operations) == 1:
            payload = operations[0]
        else:
            payload = operations
            
        headers = {
            "Content-Type": "text/plain;charset=UTF-8",
            "Client-ID": self._client_type.CLIENT_ID,
            "Authorization": f"OAuth {self.auth.access_token}",
        }
        
        # Add device ID if available
        if self.auth.device_id:
            headers["X-Device-Id"] = self.auth.device_id
            
        async with session.post(
            "https://gql.twitch.tv/gql",
            json=payload,
            headers=headers
        ) as response:
            if response.status != 200:
                raise GQLException(f"GraphQL request failed: {response.status}")
                
            json_data = await response.json()
            
            # Return as list for consistency
            if isinstance(json_data, list):
                return json_data
            else:
                return [json_data]

    def _build_gql_operation(self, operation_name: str, variables: Optional[Dict] = None) -> Dict[str, Any]:
        """Build a GraphQL operation using persisted queries"""
        if operation_name not in GQL_OPERATIONS:
            raise GQLException(f"Unknown operation: {operation_name}")
            
        operation = {
            "operationName": operation_name,
            "extensions": {
                "persistedQuery": {
                    "version": 1,
                    "sha256Hash": GQL_OPERATIONS[operation_name]
                }
            }
        }
        
        if variables:
            operation["variables"] = variables
            
        return operation

    async def get_drops_dashboard(self) -> JsonType:
        """Get the drops dashboard data"""
        operation = self._build_gql_operation("ViewerDropsDashboard", {
            "fetchRewardCampaigns": False
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def get_inventory(self) -> JsonType:
        """Get user inventory"""
        operation = self._build_gql_operation("Inventory", {
            "fetchRewardCampaigns": False
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def get_streams_for_game(self, game_name: str, limit: int = 20) -> List[Dict[str, Any]]:
        """Get streams for a specific game"""
        operation = self._build_gql_operation("DirectoryPage_Game", {
            "name": game_name,
            "options": {
                "includeRestricted": ["SUB_ONLY_LIVE"],
                "sort": "RELEVANCE",
                "systemFilters": ["DROPS_ENABLED"],
                "tags": [],
                "broadcasterLanguages": [],
                "freeformTags": None,
                "recommendationsContext": {"platform": "web"},
                "requestID": "JIRA-VXP-2397",
            },
            "sortTypeIsRecency": False,
            "limit": limit,
            "imageWidth": 50,
            "includeIsDJ": False,
        })
        
        results = await self.gql_request([operation])
        
        # Parse streams from response
        streams = []
        if results and "data" in results[0]:
            data = results[0]["data"]
            if "game" in data and data["game"] and "streams" in data["game"]:
                stream_data = data["game"]["streams"]
                if "edges" in stream_data:
                    for edge in stream_data["edges"]:
                        if "node" in edge:
                            node = edge["node"]
                            broadcaster = node.get("broadcaster", {})
                            
                            stream = {
                                "id": node.get("id", ""),
                                "user_login": broadcaster.get("login", ""),
                                "user_name": broadcaster.get("displayName", ""),
                                "title": node.get("title", ""),
                                "viewer_count": node.get("viewersCount", 0),
                                "language": broadcaster.get("broadcastSettings", {}).get("language", "en"),
                            }
                            streams.append(stream)
                            
        return streams

    async def get_stream_access_token(self, channel_login: str) -> JsonType:
        """Get stream access token (simulates watching)"""
        operation = self._build_gql_operation("PlaybackAccessToken", {
            "isLive": True,
            "isVod": False,
            "login": channel_login,
            "platform": "web",
            "playerType": "site",
            "vodID": "",
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def get_channel_points_context(self, channel_login: str) -> JsonType:
        """Get channel points context"""
        operation = self._build_gql_operation("ChannelPointsContext", {
            "channelLogin": channel_login
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def claim_drop_reward(self, drop_instance_id: str) -> JsonType:
        """Claim a drop reward"""
        operation = self._build_gql_operation("DropsPage_ClaimDropRewards", {
            "input": {"dropInstanceID": drop_instance_id}
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def claim_community_points(self, claim_id: str, channel_id: str) -> JsonType:
        """Claim community points"""
        operation = self._build_gql_operation("ClaimCommunityPoints", {
            "input": {
                "claimID": claim_id,
                "channelID": channel_id
            }
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def get_current_drop_context(self, channel_id: str) -> JsonType:
        """Get current drop context for a channel"""
        operation = self._build_gql_operation("DropCurrentSessionContext", {
            "channelID": channel_id,
            "channelLogin": ""  # Always empty in original
        })
        
        results = await self.gql_request([operation])
        return results[0] if results else {}

    async def save_cookies(self):
        """Save cookies to file"""
        if self._session and self._session.cookie_jar:
            self._session.cookie_jar.save(COOKIES_PATH)

    def start_device_auth(self) -> Dict[str, Any]:
        """Start device authorization flow - synchronous using requests"""
        import requests
        
        client_info = self._client_type
        headers = {
            "Accept": "application/json",
            "Accept-Encoding": "gzip",
            "Accept-Language": "en-US",
            "Cache-Control": "no-cache",
            "Client-Id": client_info.CLIENT_ID,
            "Host": "id.twitch.tv",
            "Origin": str(client_info.CLIENT_URL),
            "Pragma": "no-cache",
            "Referer": str(client_info.CLIENT_URL),
            "User-Agent": client_info.USER_AGENT,
        }
        
        # Start device authorization flow
        payload = {
            "client_id": client_info.CLIENT_ID,
            "scopes": "",
        }
        
        try:
            response = requests.post(
                "https://id.twitch.tv/oauth2/device",
                headers=headers,
                data=payload,
                timeout=10
            )
            response.raise_for_status()
            json_data = response.json()
            
            return {
                "verification_uri": json_data["verification_uri"],
                "user_code": json_data["user_code"], 
                "device_code": json_data["device_code"],
                "expires_in": json_data["expires_in"],
                "interval": json_data["interval"],
            }
        except Exception as e:
            logger.error(f"Device auth start error: {e}")
            raise

    def poll_device_auth(self, device_code: str, interval: int) -> Optional[str]:
        """Poll for device authorization completion - synchronous using requests"""
        import requests
        import time
        
        client_info = self._client_type
        headers = {
            "Accept": "application/json",
            "Client-Id": client_info.CLIENT_ID,
        }
        
        payload = {
            "client_id": client_info.CLIENT_ID,
            "device_code": device_code,
            "grant_type": "urn:ietf:params:oauth:grant-type:device_code",
        }
        
        # Poll for up to 5 minutes
        max_attempts = 300 // interval
        for attempt in range(max_attempts):
            logger.info(f"Token polling attempt {attempt + 1}/{max_attempts}")
            time.sleep(interval)
            
            try:
                response = requests.post(
                    "https://id.twitch.tv/oauth2/token",
                    headers=headers,
                    data=payload,
                    timeout=10
                )
                
                if response.status_code == 200:
                    json_data = response.json()
                    access_token = json_data["access_token"]
                    self.auth.access_token = access_token
                    logger.info("Successfully obtained access token!")
                    return access_token
                elif response.status_code == 400:
                    try:
                        json_data = response.json()
                        error = json_data.get("error", "")
                        logger.info(f"Polling status: {error}")
                        if error == "authorization_pending":
                            continue
                        elif error in ("expired_token", "access_denied"):
                            logger.warning(f"Auth polling failed: {error}")
                            break
                        elif error == "slow_down":
                            logger.info("Rate limited, waiting longer...")
                            time.sleep(interval)
                            continue
                    except:
                        logger.warning("Could not parse error response")
                        continue
                else:
                    logger.warning(f"Unexpected response status: {response.status_code}")
                    continue
                    
            except Exception as e:
                logger.error(f"Token polling error: {e}")
                continue
                
        logger.warning("Token polling completed without success")
        return None