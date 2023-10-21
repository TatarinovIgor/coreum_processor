INSERT INTO merchants (created_at, updated_at, key, value, ttl)
    VALUES (now(), now(),
            'aaef4567-b438-48a4-9a3a-f3a730b0e1ec',
            '{' ||
                '"id":"aaef4567-b438-48a4-9a3a-f3a730b0e1ec",' ||
                '"public_key":"-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwaT1yczs2t6ejQXPjMkX\nBEhArYEaolcHjjw0dNHBuLFQd5s1epXvS7Mx67p26KE724AG7Y4b5EC7In+MyTTS\nPY5X5/FLnRTJMRuywZHjwFUgTywP2Mdl5VbwV8bhjTUAd66ImM5YHcQEeYGgXhIb\nADk0qYhwp47R5QEm9ID0bE3gDc1dvm+Ak45pZBKBAL1AJptb6jzTdwpR0X9yS2g0\n4z0MoRcDjFoB9f0pj4XZiuH1TJzIn/IGkRKKfCht1bRh3UX7KESGkeHs+LW022+u\nYGK3m/Nor1fd5apLEaCFDjjrgYmr4YijzxPtPf/aqVMKAZMW+4p75C35KFspiOmh\nmDr5mDN+2g+YOu2wBGj9glvKFJFfWAVB3DbEZjLDGN4fVJx4wKAkdeL8I0j+cQWA\nZY8J8Q8KN06u7aWN8BmoaRF/KdcPpMBMlV9T8e5M7bgruYUifg4AnvICsJG4adFA\n9bEiT8/PIpnhDHGQGTcS1HvZPxVCcbFjX7ePG1q3wvr69hracPn7YgNT0af6xJ8g\nHCusgppRJ3Rh6YuCUFa2kY0GzeNy9Jk6zJrqW+SwFz45xXVgY0mocYervEzpi3Dj\nyoI2UAFJgyu/xL28iZ0v22gGEfEkzF5SRePwtCUhmwjABhfnNXHfxMsYI+jI95Y0\nUd/f2KZt2uHJ7i3MsYyjCoECAwEAAQ==\n-----END PUBLIC KEY-----", ' ||
                '"name":"Merchant",' ||
                '"call_back_url":"",' ||
                '"wallets":{' ||
                    '"coreum":{' ||
                        '"commission_receiving":{"fix":1,"percent":1},' ||
                        '"commission_sending":{"fix":1,"percent":1},' ||
                        '"receiving_id":"aaef4567-b438-48a4-9a3a-f3a730b0e1ec-R",' ||
                        '"sending_id":"aaef4567-b438-48a4-9a3a-f3a730b0e1ec-S"' ||
                    '}' ||
                '}' ||
            '}',
            9223372036854775807)
    ON CONFLICT (key) do update set updated_at = excluded.updated_at,
                                    value = excluded.value;
