import decimal
from abc import ABC, abstractmethod
from collections.abc import Mapping
from typing import Optional

import gevent.socket
from pydantic import ConfigDict

from dify_plugin.entities import I18nObject
from dify_plugin.entities.model import (
    PARAMETER_RULE_TEMPLATE,
    AIModelEntity,
    DefaultParameterName,
    ModelType,
    PriceConfig,
    PriceInfo,
    PriceType,
)
from dify_plugin.errors.model import InvokeAuthorizationError, InvokeError

import socket

if socket.socket is gevent.socket.socket:
    import gevent.threadpool

    threadpool = gevent.threadpool.ThreadPool(1)


class AIModel(ABC):
    """
    Base class for all models.
    """

    model_type: ModelType
    model_schemas: list[AIModelEntity]
    started_at: float = 0

    # pydantic configs
    model_config = ConfigDict(protected_namespaces=())

    def __init__(self, model_schemas: list[AIModelEntity]) -> None:
        self.model_schemas = [
            model_schema for model_schema in model_schemas if model_schema.model_type == self.model_type
        ]

    @abstractmethod
    def validate_credentials(self, model: str, credentials: Mapping) -> None:
        """
        Validate model credentials

        :param model: model name
        :param credentials: model credentials
        :return:
        """
        raise NotImplementedError

    @property
    @abstractmethod
    def _invoke_error_mapping(self) -> dict[type[InvokeError], list[type[Exception]]]:
        """
        Map model invoke error to unified error
        The key is the error type thrown to the caller
        The value is the error type thrown by the model,
        which needs to be converted into a unified error type for the caller.

        :return: Invoke error mapping
        """
        raise NotImplementedError

    def _transform_invoke_error(self, error: Exception) -> InvokeError:
        """
        Transform invoke error to unified error

        :param error: model invoke error
        :return: unified error
        """
        provider_name = self.__class__.__module__.split(".")[-3]

        for invoke_error, model_errors in self._invoke_error_mapping.items():
            if isinstance(error, tuple(model_errors)):
                if invoke_error == InvokeAuthorizationError:
                    return invoke_error(
                        description=f"[{provider_name}] Incorrect model credentials provided, "
                        "please check and try again. "
                    )

                return invoke_error(description=f"[{provider_name}] {invoke_error.description}, {str(error)}")

        return InvokeError(description=f"[{provider_name}] Error: {str(error)}")

    def get_price(self, model: str, credentials: dict, price_type: PriceType, tokens: int) -> PriceInfo:
        """
        Get price for given model and tokens

        :param model: model name
        :param credentials: model credentials
        :param price_type: price type
        :param tokens: number of tokens
        :return: price info
        """
        # get model schema
        model_schema = self.get_model_schema(model, credentials)

        # get price info from predefined model schema
        price_config: Optional[PriceConfig] = None
        if model_schema and model_schema.pricing:
            price_config = model_schema.pricing

        # get unit price
        unit_price = None
        if price_config:
            if price_type == PriceType.INPUT:
                unit_price = price_config.input
            elif price_type == PriceType.OUTPUT and price_config.output is not None:
                unit_price = price_config.output

        if unit_price is None:
            return PriceInfo(
                unit_price=decimal.Decimal("0.0"),
                unit=decimal.Decimal("0.0"),
                total_amount=decimal.Decimal("0.0"),
                currency="USD",
            )

        # calculate total amount
        if not price_config:
            raise ValueError(f"Price config not found for model {model}")
        total_amount = tokens * unit_price * price_config.unit
        total_amount = total_amount.quantize(decimal.Decimal("0.0000001"), rounding=decimal.ROUND_HALF_UP)

        return PriceInfo(
            unit_price=unit_price,
            unit=price_config.unit,
            total_amount=total_amount,
            currency=price_config.currency,
        )

    def predefined_models(self) -> list[AIModelEntity]:
        """
        Get all predefined models for given provider.

        :return:
        """
        return self.model_schemas

    def get_model_schema(self, model: str, credentials: Optional[Mapping] = None) -> Optional[AIModelEntity]:
        """
        Get model schema by model name and credentials

        :param model: model name
        :param credentials: model credentials
        :return: model schema
        """
        # get predefined models (predefined_models)
        models = self.predefined_models()

        model_map = {model.model: model for model in models}
        if model in model_map:
            return model_map[model]

        if credentials:
            model_schema = self.get_customizable_model_schema_from_credentials(model, credentials)
            if model_schema:
                return model_schema

        return None

    def get_customizable_model_schema_from_credentials(
        self, model: str, credentials: Mapping
    ) -> Optional[AIModelEntity]:
        """
        Get customizable model schema from credentials

        :param model: model name
        :param credentials: model credentials
        :return: model schema
        """
        return self._get_customizable_model_schema(model, credentials)

    def _get_customizable_model_schema(self, model: str, credentials: Mapping) -> Optional[AIModelEntity]:
        """
        Get customizable model schema and fill in the template
        """
        schema = self.get_customizable_model_schema(model, credentials)

        if not schema:
            return None

        # fill in the template
        new_parameter_rules = []
        for parameter_rule in schema.parameter_rules:
            if parameter_rule.use_template:
                try:
                    default_parameter_name = DefaultParameterName.value_of(parameter_rule.use_template)
                    default_parameter_rule = self._get_default_parameter_rule_variable_map(default_parameter_name)
                    if not parameter_rule.max and "max" in default_parameter_rule:
                        parameter_rule.max = default_parameter_rule["max"]
                    if not parameter_rule.min and "min" in default_parameter_rule:
                        parameter_rule.min = default_parameter_rule["min"]
                    if not parameter_rule.default and "default" in default_parameter_rule:
                        parameter_rule.default = default_parameter_rule["default"]
                    if not parameter_rule.precision and "precision" in default_parameter_rule:
                        parameter_rule.precision = default_parameter_rule["precision"]
                    if not parameter_rule.required and "required" in default_parameter_rule:
                        parameter_rule.required = default_parameter_rule["required"]
                    if not parameter_rule.help and "help" in default_parameter_rule:
                        parameter_rule.help = I18nObject(
                            en_US=default_parameter_rule["help"]["en_US"],
                        )
                    if (
                        parameter_rule.help
                        and not parameter_rule.help.en_US
                        and ("help" in default_parameter_rule and "en_US" in default_parameter_rule["help"])
                    ):
                        parameter_rule.help.en_US = default_parameter_rule["help"]["en_US"]
                    if (
                        parameter_rule.help
                        and not parameter_rule.help.zh_Hans
                        and ("help" in default_parameter_rule and "zh_Hans" in default_parameter_rule["help"])
                    ):
                        parameter_rule.help.zh_Hans = default_parameter_rule["help"].get(
                            "zh_Hans", default_parameter_rule["help"]["en_US"]
                        )
                except ValueError:
                    pass

            new_parameter_rules.append(parameter_rule)

        schema.parameter_rules = new_parameter_rules

        return schema

    def get_customizable_model_schema(self, model: str, credentials: Mapping) -> Optional[AIModelEntity]:
        """
        Get customizable model schema

        :param model: model name
        :param credentials: model credentials
        :return: model schema
        """
        return None

    def _get_default_parameter_rule_variable_map(self, name: DefaultParameterName) -> dict:
        """
        Get default parameter rule for given name

        :param name: parameter name
        :return: parameter rule
        """
        default_parameter_rule = PARAMETER_RULE_TEMPLATE.get(name)

        if not default_parameter_rule:
            raise Exception(f"Invalid model parameter rule name {name}")

        return default_parameter_rule

    def _get_num_tokens_by_gpt2(self, text: str) -> int:
        """
        Get number of tokens for given prompt messages by gpt2
        Some provider models do not provide an interface for obtaining the number of tokens.
        Here, the gpt2 tokenizer is used to calculate the number of tokens.
        This method can be executed offline, and the gpt2 tokenizer has been cached in the project.

        :param text: plain text of prompt. You need to convert the original message to plain text
        :return: number of tokens
        """

        # ENHANCEMENT:
        # to avoid performance issue, do not calculate the number of tokens for too long text
        # only to promise text length is less than 100000
        if len(text) >= 100000:
            return len(text)

        import tiktoken

        # check if gevent is patched to main thread
        import socket

        if socket.socket is gevent.socket.socket:
            # using gevent real thread to avoid blocking main thread
            result = threadpool.spawn(lambda: len(tiktoken.encoding_for_model("gpt2").encode(text)))
            return result.get(block=True) or 0

        return len(tiktoken.encoding_for_model("gpt2").encode(text))
