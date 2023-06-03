FROM python:3.7-slim

WORKDIR /app

ADD . .

RUN pip install -i https://pypi.tuna.tsinghua.edu.cn/simple/ -r requirements.txt

EXPOSE 9090

CMD ["python", "./main.py"]