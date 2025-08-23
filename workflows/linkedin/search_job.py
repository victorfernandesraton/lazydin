import logging
import os
from urllib.parse import urlparse

from lazydin.browser import NoDriverService
import nodriver as uc


class LinkedinSearch:
    """Implement a way to search on LinkedIn using an existing browser session"""
    SEARCH_INPUT = "//input[@placeholder='Search']"
    browser_service: NoDriverService
    linkedin_domain: str

    def __init__(
        self,
        browser_service: NoDriverService = None,
        linkedin_domain="https://www.linkedin.com",
    ):
        self.browser_service = browser_service
        self.linkedin_domain = linkedin_domain

    async def execute(self, params: dict = None):
        if params is None:
            params = {}

        search = params.get("search", "")
        driver_key = params.get("driver_key")
        username = params.get("username", "default_user")
        # Store keep_browser_open flag - default to True
        keep_browser_open = params.get("keep_browser_open", True)
        use_cookies = params.get("use_cookies", True)
        # category = params.get("category", "Post")

        if self.browser_service is None:
            self.browser_service = NoDriverService()
            driver_key = await self.browser_service.open_browser()
            # Store driver_key in params to pass it to subsequent tasks
            params["driver_key"] = driver_key

        logging.info("go to search")
        driver = self.browser_service.drivers[driver_key]
        
        # Create a cookie filename based on the username
        cookie_file = f"linkedin_{username.replace('@', '_at_')}.session"
        
        # Try to load cookies if use_cookies is True and we don't already have an active session
        cookies_loaded = False
        if use_cookies and not params.get("cookies_saved") and os.path.exists(cookie_file):
            try:
                logging.info(f"Loading cookies from {cookie_file}")
                # First navigate to the domain to be able to set cookies
                page = await driver.get(self.linkedin_domain)
                
                # Load cookies directly using the browser's built-in cookie functionality
                await driver.cookies.load(cookie_file)
                logging.info("Cookies loaded successfully")
                cookies_loaded = True
            except Exception as e:
                logging.warning(f"Error loading cookies: {e}")
                cookies_loaded = False
        
        # Go directly to feed page
        page = await driver.get(f"{self.linkedin_domain}/feed")
        await page.maximize()
        await page.find("Home", best_match=True)
        logging.info("set input")

        input_search = await page.xpath(self.SEARCH_INPUT, timeout=10)
        if not input_search:
            raise Exception(f"Not found search input {self.SEARCH_INPUT}")

        await self.browser_service.human_input_simulate(input_search[0], search)
        await page.send

        await input_search[0].send(uc.cdp.input_.dispatch_key_event(
            type_="rawKeyDown",
            windows_virtual_key_code=13,
        ))

        # Store the keep_browser_open flag - default to True
        keep_browser_open = params.get("keep_browser_open", True)
        
        # Only close the page if not keeping browser open
        if not keep_browser_open:
            await page.close()
            # Only close the browser if we're not keeping it open
            self.browser_service.close()

        return {
            "status": "success", 
            "message": "Search completed",
            "driver_key": driver_key,  # Return driver_key for potential further chaining
            "keep_browser_open": keep_browser_open  # Pass this flag along
        }
