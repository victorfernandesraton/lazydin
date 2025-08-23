import logging
import os

from lazydin.browser import NoDriverService
import nodriver as uc
from asyncio import sleep

logger = logging.getLogger(__name__)


class LinkedinSearch:
    """Implement a way to search on LinkedIn using an existing browser session"""

    SEARCH_INPUT = "//input[@placeholder='Search']"
    CATEGORY_BUTTON = "//button[text()='VALUE']"
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
        category = params.get("category", "Posts")
        driver_key = params.get("driver_key")
        username = params.get("username", "default_user")
        # Store keep_browser_open flag - default to True
        keep_browser_open = params.get("keep_browser_open", True)
        if self.browser_service is None:
            self.browser_service = NoDriverService()
            driver_key = await self.browser_service.open_browser()
            # Store driver_key in params to pass it to subsequent tasks
            params["driver_key"] = driver_key

        logger.info("go to search")
        driver = self.browser_service.drivers[driver_key]

        # Create a cookie filename based on the username
        cookie_file = f"linkedin_{username.replace('@', '_at_')}.session"
        if not os.path.exists(cookie_file):
            raise Exception(f"cookie for {username} not available")

        logger.info(f"Loading cookies from {username}")
        # First navigate to the domain to be able to set cookies
        page = await driver.get(self.linkedin_domain)

        # Load cookies directly using the browser's built-in cookie functionality
        await driver.cookies.load(cookie_file)
        logger.info("Cookies loaded successfully")

        # Go directly to feed page
        page = await driver.get(f"{self.linkedin_domain}/feed")
        await page.maximize()

        logger.info("set input")

        # Try to find the search input with retries
        max_retries = 3
        retry_count = 0
        input_search = None

        while retry_count < max_retries:
            try:
                input_search = await page.xpath(self.SEARCH_INPUT, timeout=15)
                if input_search:
                    break
                logger.warning(
                    f"Search input not found, retrying ({retry_count + 1}/{max_retries})"
                )
                retry_count += 1
                await sleep(2)  # Wait before retrying
            except Exception as e:
                logger.warning(
                    f"Error finding search input: {e}, retrying ({retry_count + 1}/{max_retries})"
                )
                retry_count += 1
                await sleep(2)  # Wait before retrying

        if not input_search:
            raise Exception(
                f"Not found search input {self.SEARCH_INPUT} after {max_retries} attempts"
            )

        await self.browser_service.human_input_simulate(input_search[0], search)

        # Send Enter key using the correct method
        await page.send(
            uc.cdp.input_.dispatch_key_event(
                type_="keyDown",
                key="Enter",
                code="Enter",
                windows_virtual_key_code=13,
            )
        )

        logger.info(f"Get button to category {category}")

        button_category = await page.xpath(
            self.CATEGORY_BUTTON.replace("VALUE", category)
        )

        if not button_category:
            raise Exception(f"button_category {category} not found")

        await button_category[0].click()

        return {
            "status": "success",
            "message": "Search completed",
            "driver_key": driver_key,  # Return driver_key for potential further chaining
        }
