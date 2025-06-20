import pandas as pd
import random
import string

def random_string(prefix, length=6):
    return prefix + ''.join(random.choices(string.ascii_uppercase + string.digits, k=length))

servers = []
for i in range(7, 11):
    servers.append({
        "server_id": random_string("SRV"),
        "server_name": random_string("Server"),
        "status": random.choice(["ON"]),
        "ipv4": f"192.168.100.{i}",
        "description": f"Generated server {i}",
        "location": random.choice(["US-East", "US-West", "EU-Central", "Asia-Pacific"]),
        "os": random.choice(["Ubuntu 20.04", "CentOS 7", "Debian 11", "Windows Server 2019"]),
        "cpu": random.choice([2, 4, 8, 16]),
        "ram": random.choice([4, 8, 16, 32]),
        "disk": random.choice([100, 250, 500, 1000]),
    })

# Ghi vào file Excel
df = pd.DataFrame(servers)
df.to_excel("servers_10000.xlsx", index=False)
print("✅ Đã tạo file servers_10000.xlsx")
