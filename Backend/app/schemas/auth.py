from pydantic import BaseModel, EmailStr, validator
from Backend.utils.validation_utils import validate_phone_number


class Token(BaseModel):
    access_token: str
    token_type: str


class TokenData(BaseModel):
    email: str | None = None


class UserCreate(BaseModel):
    username: str
    email: EmailStr
    password: str
    first_name: str | None = None
    last_name: str | None = None
    phone_number: str | None = None

    @validator('phone_number')
    def validate_phone(cls, v):
        if v is not None:
            result = validate_phone_number(v)
            if not result["is_valid"]:
                raise ValueError(
                    f"Invalid phone number. Requirements: {result['requirements']}")
            return result["formatted_number"]
        return v
