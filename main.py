import asyncio
import logging
import re
from aiogram import Bot, Dispatcher, types
from aiogram.filters import CommandStart, Command
from aiogram.methods import DeleteWebhook
from aiogram.types import Message, InlineKeyboardMarkup, InlineKeyboardButton
import requests
import json
import os

TOKEN = '7649469697:AAGKM8xOmG62Xg3_YABzKux7Lrk-IW60c9w'  # Ваш токен
CHANNEL_ID = '-1002646685313'  # Замените на username или ID вашего канала
LINK = "https://t.me/ilyapanov_yt"
logging.basicConfig(level=logging.INFO)
bot = Bot(TOKEN)
dp = Dispatcher()

user_models = {}

models = {
    "Qwen-32B": "Qwen/QwQ-32B",
    "Llama-3.2": "meta-llama/Llama-3.2-90B-Vision-Instruct",
    "DeepSeek-R1": "deepseek-ai/DeepSeek-R1",
    "Llama-3.3": "meta-llama/Llama-3.3-70B-Instruct",
    "dbrx-instruct": "databricks/dbrx-instruct",
    "Ministral-8B": "mistralai/Ministral-8B-Instruct-2410",
    "Confucius-14B": "netease-youdao/Confucius-o1-14B",
    "AceMath-7B": "nvidia/AceMath-7B-Instruct",
    "gemma-2-9B": "google/gemma-2-9b-it",
    "Llama-3.1": "neuralmagic/Llama-3.1-Nemotron-70B-Instruct-HF-FP8-dynamic",
    "Mistral-Large": "mistralai/Mistral-Large-Instruct-2411",
    "phi-4": "microsoft/phi-4",
    "watt-tool": "watt-ai/watt-tool-70B",
    "Bespoke-32B": "bespokelabs/Bespoke-Stratos-32B",
    "Sky-T1": "NovaSky-AI/Sky-T1-32B-Preview",
    "Falcon3-10B": "tiiuae/Falcon3-10B-Instruct",
    "c4ai-command": "CohereForAI/c4ai-command-r-plus-08-2024",
    "glm-4-9B": "THUDM/glm-4-9b-chat",
    "Qwen2.5-Coder": "Qwen/Qwen2.5-Coder-32B-Instruct",
    "aya-expanse": "CohereForAI/aya-expanse-32b",
    "ReaderLM-v2": "jinaai/ReaderLM-v2",
    "MiniCPM3-4B": "openbmb/MiniCPM3-4B",
    "Qwen2.5-1.5B": "Qwen/Qwen2.5-1.5B-Instruct",
    "0x-lite": "ozone-ai/0x-lite",
    "Phi-3.5-mini": "microsoft/Phi-3.5-mini-instruct"
}

model_descriptions = {
    "Qwen-32B": "Модель Qwen-32B с передовыми возможностями обработки языка для исследований.",
    "Llama-3.2": "Модель Llama-3.2 с поддержкой визуальных инструкций.",
    "DeepSeek-R1": "Высокопроизводительная модель для генерации текста, суммирования и выполнения инструкций.",
    "Llama-3.3": "Крупномасштабная модель, дообученная для точного выполнения инструкций.",
    "dbrx-instruct": "Модель для точных, задач-ориентированных ответов, идеально подходит для корпоративных приложений.",
    "Ministral-8B": "Мощная модель для создания качественного текста и выполнения инструкций.",
    "Confucius-14B": "Модель для качественной обработки текстов.",
    "AceMath-7B": "Специализированная модель для математического мышления и решения задач.",
    "gemma-2-9B": "Лёгкая, но мощная модель для эффективных и контекстно-зависимых ответов.",
    "Llama-3.1": "Модель для точного выполнения инструкций с динамической точностью.",
    "Mistral-Large": "Мощная языковая модель для глубокого понимания инструкций.",
    "phi-4": "Компактная и аналитичная модель с выдающимися способностями.",
    "watt-tool": "Модель для решения сложных задач.",
    "Bespoke-32B": "Специализированная модель для уникальных задач.",
    "Sky-T1": "Предварительная версия с передовыми возможностями обработки данных.",
    "Falcon3-10B": "Надёжная модель для выполнения инструкций.",
    "c4ai-command": "Модель для корпоративного стиля ответов.",
    "glm-4-9B": "Модель для интерактивного общения.",
    "Qwen2.5-Coder": "Модель для генерации кода и программирования.",
    "aya-expanse": "Модель для качественной обработки текстов.",
    "ReaderLM-v2": "Модель для глубокого анализа текстов.",
    "MiniCPM3-4B": "Модель для генерации компактных и точных текстов.",
    "Qwen2.5-1.5B": "Лёгкая версия для выполнения инструкций.",
    "0x-lite": "Модель для быстрого решения текстовых задач.",
    "Phi-3.5-mini": "Компактная и эффективная модель для выполнения инструкций."
}

default_models = {
    "Qwen-32B": "Qwen/QwQ-32B",
    "DeepSeek-R1": "deepseek-ai/DeepSeek-R1",
    "Ministral-8B": "mistralai/Ministral-8B-Instruct-2410",
    "Mistral-Large": "mistralai/Mistral-Large-Instruct-2411",
    "Llama-3.3": "meta-llama/Llama-3.3-70B-Instruct"
}

def load_users_from_json(filename="users.json"):
    try:
        with open(filename, "r", encoding="utf-8") as f:
            return json.load(f)
    except (FileNotFoundError, json.decoder.JSONDecodeError):
        return []

def save_users_to_json(users, filename="users.json"):
    with open(filename, "w", encoding="utf-8") as f:
        json.dump(users, f, ensure_ascii=False, indent=4)

def get_full_models_keyboard():
    buttons = [
        InlineKeyboardButton(text=short_name, callback_data=full_model)
        for short_name, full_model in models.items()
    ]
    keyboard_buttons = [buttons[i:i+3] for i in range(0, len(buttons), 3)]
    return InlineKeyboardMarkup(inline_keyboard=keyboard_buttons)

def get_default_models_keyboard():
    buttons = [
        InlineKeyboardButton(text=short_name, callback_data=full_model)
        for short_name, full_model in default_models.items()
    ]
    buttons.append(InlineKeyboardButton(text="Больше моделей", callback_data="more_models"))
    keyboard_buttons = [buttons[i:i+3] for i in range(0, len(buttons), 3)]
    return InlineKeyboardMarkup(inline_keyboard=keyboard_buttons)

def get_start_keyboard():
    keyboard = InlineKeyboardMarkup(inline_keyboard=[
        [InlineKeyboardButton(text="О моделях", callback_data="about_models")],
        [InlineKeyboardButton(text="Выбрать модель", callback_data="select_model")]
    ])
    return keyboard

# ----------- ФУНКЦИЯ ПРОВЕРКИ ПОДПИСКИ -----------
async def check_subscription(user_id):
    try:
        member = await bot.get_chat_member(chat_id=CHANNEL_ID, user_id=user_id)
        return member.status in ('member', 'administrator', 'creator')
    except Exception:
        return False

# ----------- ДЕКОРАТОР ДЛЯ ПРОВЕРКИ ПОДПИСКИ -----------
def require_subscription(handler):
    async def wrapper(message: Message, *args, **kwargs):
        if not await check_subscription(message.from_user.id):
            markup = InlineKeyboardMarkup(inline_keyboard=[
                [InlineKeyboardButton(text="✅ Подписаться на канал", url=LINK)],
            ])
            await message.answer(
                "Чтобы пользоваться ботом, подпишитесь на канал и нажмите /start ещё раз!",
                reply_markup=markup
            )
            return
        return await handler(message, *args, **kwargs)  # Передаём все аргументы
    return wrapper

# ----------- ДЕКОРАТОР ДЛЯ CALLBACK (для aiogram 3) -----------
def require_subscription_callback(handler):
    async def wrapper(callback: types.CallbackQuery, *args, **kwargs):
        if not await check_subscription(callback.from_user.id):
            markup = InlineKeyboardMarkup(inline_keyboard=[
                [InlineKeyboardButton(text="✅ Подписаться на канал",url=LINK)]
            ])
            await callback.message.answer(
                "Чтобы пользоваться ботом, подпишитесь на наш канал и нажмите /start ещё раз!",
                reply_markup=markup
            )
            await callback.answer()
            return
        return await handler(callback, *args, **kwargs)
    return wrapper

# ----------- ОБРАБОТЧИКИ С ПРОВЕРКОЙ ПОДПИСКИ -----------

@dp.message(Command("start"))
@require_subscription
async def cmd_start(message: types.Message, **kwargs):
    if not os.path.exists("users.json"):
        with open("users.json", "w", encoding="utf-8") as f:
            json.dump([], f)
    user = message.from_user
    users = load_users_from_json()
    
    # Проверяем, есть ли пользователь в списке
    user_exists = False
    for u in users:
        if u["user_id"] == user.id:
            user_exists = True
            # Обновляем данные, если изменились имя/фамилия/username
            u.update({
                "first_name": user.first_name,
                "last_name": user.last_name,
                "username": user.username
            })
            break
    
    if not user_exists:
        users.append({
            "user_id": user.id,
            "first_name": user.first_name,
            "last_name": user.last_name,
            "username": user.username
        })
    
    save_users_to_json(users)
    
    keyboard = get_start_keyboard()
    await message.answer(
        "Добро пожаловать в волшебный мир нейросети! Выберите интересующую Вас опцию ниже:",
        reply_markup=keyboard,
        parse_mode="HTML"
    )

@dp.message(Command("models"))
@require_subscription
async def cmd_model(message: Message, **kwargs):
    keyboard = get_default_models_keyboard()
    await message.answer("Пожалуйста, выберите модель, которая Вас интересует:", reply_markup=keyboard)

@dp.callback_query()
@require_subscription_callback
async def handle_callback(callback: types.CallbackQuery, **kwargs):
    data = callback.data
    if data == "about_models":
        text_lines = ["<b>Доступные модели:</b>"]
        for short_name, full_model in models.items():
            description = model_descriptions.get(short_name, "Описание отсутствует.")
            text_lines.append(f"• <b>{short_name}</b> ({full_model})\n  <i>{description}</i>")
        await callback.message.answer("\n".join(text_lines), parse_mode="HTML")
        await callback.answer()
    elif data == "select_model":
        keyboard = get_default_models_keyboard()
        await callback.message.answer("Выберите модель из списка:", reply_markup=keyboard)
        await callback.answer()
    elif data == "more_models":
        keyboard = get_full_models_keyboard()
        await callback.message.answer("Пожалуйста, выберите модель из полного списка:", reply_markup=keyboard)
        await callback.answer()
    else:
        selected_model = data
        short_name = next((k for k, v in models.items() if v == selected_model), selected_model)
        user_models[callback.from_user.id] = selected_model
        await callback.answer(f"Модель выбрана: {short_name}. Теперь напишите Ваш запрос.", show_alert=True)

@dp.message()
@require_subscription
async def filter_messages(message: Message, **kwargs):
    await message.answer("Ваш запрос обрабатывается, пожалуйста, подождите...")
    model_used = user_models.get(message.from_user.id, "mistralai/Ministral-8B-Instruct-2410")
    url = "https://api.intelligence.io.solutions/api/v1/chat/completions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": ("Bearer io-v2-eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9."
                          "eyJvd25lciI6IjQwM2JlNGU5LWJhNDQtNGMzZC1hNjcxLTQ5OTY1NDQ2NDZmNyIsImV4cCI6NDg5NTY1Nzg0NX0."
                          "LzxmuPMQDMwRFFz11vtmtM0nitM7iiD1DqZi1XogfLiN9YwQv4JNtL3lz0MJiYLa_YcFRDYUKS56iLtHVA9L0g")
    }
    data = {
        "model": model_used,
        "messages": [
            {
                "role": "system",
                "content": "You are the best assistant and you answer perfectly in Russian."
            },
            {
                "role": "user",
                "content": message.text
            }
        ],
    }
    response = requests.post(url, headers=headers, json=data)
    data_response = response.json()
    bot_text = data_response['choices'][0]['message']['content']
    lower_text = bot_text.lower()
    if "</think>" in lower_text:
        idx = lower_text.rfind("</think>")
        bot_text = bot_text[idx + len("</think>"):]
    bot_text = re.sub(r'</?think>\n?', '', bot_text, flags=re.IGNORECASE).strip()
    max_length = 4000
    if len(bot_text) > max_length:
        for i in range(0, len(bot_text), max_length):
            await message.answer(bot_text[i:i+max_length], parse_mode="Markdown")
    else:
        await message.answer(bot_text, parse_mode="Markdown")

async def main():
    await bot(DeleteWebhook(drop_pending_updates=True))
    await dp.start_polling(bot)

if __name__ == "__main__":
    asyncio.run(main())
