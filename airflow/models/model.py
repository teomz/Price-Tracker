from pydantic import BaseModel, Field, HttpUrl
from datetime import datetime

class Omnibus(BaseModel):
    upc: str = Field(..., description="Universal Product Code")
    name: str = Field(..., description="Name of the omnibus")
    price: float = Field(..., description="Price of the omnibus")
    version: str = Field(..., description="Standard or DM version")
    pagecount: int = Field(..., description="Total number of pages")
    releaseddate: str = Field(..., description="Creation date (YYYY-MM-DD)")
    publisher: str = Field(..., description="Publisher of the omnibus")
    imgpath: str = Field(..., description="Path to the image file")
    isturl: str = Field(..., description="URL to IST")
    amazonurl: str = Field(..., description="URL to Amazon")
    last_updated: datetime = Field(default_factory=datetime.now, description="Last update timestamp")
    status: str = Field(default='Hot', description="Hot, Cold, Archive")
