import logging
from urllib.parse import urlparse
import os

from lazydin.browser import NoDriverService

logger = logging.getLogger(__name__)

class LinkedinAuth:
    """Implement a way to auth in linkedin using only username and password"""
    USERNAME_INPUT_XPATH = "//input[@id='username']"
    PASSWORD_INPUT_XPATH = "//input[@id='password']"
    BUTTON_SUBMIT_XPATH = "//button[@type='submit']"
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

        driver_key = params.get("driver_key")
        username = params.get("username")
        password = params.get("password")
        self.keep_browser_open = params.get("keep_browser_open", True)  # Default to True
        use_cookies = params.get("use_cookies", True)  # Default to using cookies
        
        # Create a cookie filename based on the username
        cookie_file = f"linkedin_{username.replace('@', '_at_')}.session"

        if self.browser_service is None:
            self.browser_service = NoDriverService()
            driver_key = await self.browser_service.open_browser()
        
        # Store driver_key in params to pass it to subsequent tasks
        params["driver_key"] = driver_key
        driver = self.browser_service.drivers[driver_key]
        
        # Try to load cookies if use_cookies is True
        cookies_loaded = False
        if use_cookies and os.path.exists(cookie_file):
            try:
                logger.info(f"Loading cookies from {cookie_file}")
                # First navigate to the domain to be able to set cookies
                page = await driver.get(self.linkedin_domain)
                
                # Load cookies directly using the browser's built-in cookie functionality
                await driver.cookies.load(cookie_file)
                logger.info(f"Cookies loaded from {cookie_file}")
                
                # Refresh to apply cookies
                page = await driver.get(f"{self.linkedin_domain}/feed")
                # Check if we're logged in
                try:
                    # Look for elements that indicate we're logged in
                    feed_elements = await page.find("Home", best_match=True, timeout=5)
                    if feed_elements:
                        logger.info("Successfully logged in using cookies")
                        cookies_loaded = True
                    else:
                        cookies_loaded = False
                except Exception as e:
                    logger.warning(f"Cookie login failed: {e}")
                    cookies_loaded = False
            except Exception as e:
                logger.warning(f"Error loading cookies: {e}")
                cookies_loaded = False
        
        # If cookies didn't work, do normal login
        if not cookies_loaded:
            logger.info("Performing normal login")
            page = await driver.get(f"{self.linkedin_domain}/login")
        await page.maximize()
        await page.find("Sign in", best_match=True)
        logger.info("set input")
        input_list = {
            self.USERNAME_INPUT_XPATH: username,
            self.PASSWORD_INPUT_XPATH: password,
        }

        for xpath, value in input_list.items():
            elements = await page.xpath(xpath, timeout=10)
            if not elements:
                raise Exception(f"Not found {xpath} input")
            await self.browser_service.human_input_simulate(
                element=elements[0],
                content=value,
                delay=0.25
            )

        btn_submit = await page.xpath(xpath=self.BUTTON_SUBMIT_XPATH, timeout=10)
        if not btn_submit:
            raise Exception("Not found button")

        await btn_submit[0].click()
        logger.info("button clicked")
        
        # Wait for login to complete and save cookies
        try:
            # Wait for redirect to feed or home page
            await page.wait_for(text="Home", timeout=30)
            
            # Save cookies for future use
            if use_cookies:
                # Save cookies directly using the browser's built-in cookie functionality
                await driver.cookies.save(cookie_file)
                logger.info(f"Saved cookies to {cookie_file}")
        except Exception as e:
            logger.warning(f"Error saving cookies: {e}")
        
        return {
            "status": "success", 
            "message": "Authentication completed",
            "driver_key": driver_key,  # Return driver_key for chaining workflows
            "keep_browser_open": self.keep_browser_open,  # Pass this flag along
            "cookies_saved": True
        }
