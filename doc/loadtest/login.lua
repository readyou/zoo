require('io')

local thread_cnt = 0
local threads = {}

function setup(thread)
    thread:set("id", thread_cnt)
    table.insert(threads, thread)
    thread_cnt = thread_cnt + 1
end

function init(args)
    -- sampled = false
    tokens = {}
    refresh_tokens = {}
    lines = {}
    lines_count = 0
    for line in io.lines('/opt/data/user.sql') do
      lines[#lines + 1] = line
      lines_count = lines_count + 1
    end

    req_count = 0
    errors = {}
    local msg = "thread %d started"
    print(msg:format(id))
end

function request()
    req_count = req_count + 1
    wrk.method = 'POST'
    wrk.headers['content-type'] = 'application/json'
    wrk.body = string.format("{\"username\":\"%s\",\"password\":\"kdjfkdjfkjdkfjdkj\"}", lines[req_count % lines_count])
    -- print(wrk.body)
    -- return wrk.request()
    return wrk.format()
end

function format(data)
    local text = ''
    if type(data) == 'table' then
        text = text .. '{'
        local is_first = true
        for key, value in pairs(data) do
            if is_first then
                is_first = false
            else
                text = text .. ','
            end
            text = text .. tostring(key) .. ':' .. tostring(value)
        end
        text = text .. '}'
    else
        text = tostring(data)
    end
    return text
end

function response(status, headers, body)
    if status == 200 then
        if string.find(body, '"Success":false') then
            -- _, _, code = string.find(body, '"code":"(%d+)"')
            _, _, code = string.find(body, '"Code":"(%w+)"')
            if errors[code] then
                errors[code] = errors[code] + 1
            else
                errors[code] = 1
            end
        else
            _, _, token = string.find(body, '"Token":"(%S-)"')
            _, _, refresh = string.find(body, '"RefreshToken":"(%S-)"')
            table.insert(tokens, token)
            table.insert(refresh_tokens, refresh)
        end
    else
        local cnt = errors[status]
        if cnt then
            errors[status] = cnt + 1
        else
            errors[status] = 1
        end
    end

    -- if not sampled then
    --     sampled = true
    --     print(body)
    -- end

    if req_count % 1000 == 0 then
        local msg = "thread: %d, total: %d, errors: %s"
        print(msg:format(id, req_count, format(errors)))
    end
end


function done(summary, latency, requests)
    print("\nAdditional Info")
    local req_total = 0
    local error_total = 0
    local errors_all = {}
    local ftoken = io.open('/opt/data/token.txt', 'w+')
    local frefresh = io.open('/opt/data/refresh.txt', 'w+')
    for index, thread in ipairs(threads) do
        local id = thread:get("id")
        local req_count = thread:get("req_count")
        req_total = req_total + req_count
        local errors = thread:get("errors")
        for k, v in pairs(errors) do
            error_total = error_total + v
            if errors_all[k] then
                errors_all[k] = errors_all[k] + v
            else
                errors_all[k] = v
            end
        end
        local msg = "thread %d, req_count: %d, errors: %s"
        print(msg:format(id, req_count, format(errors)))

        local tokens = thread:get("tokens")
        local refresh_tokens = thread:get("refresh_tokens")
        for i, v in ipairs(tokens) do
          ftoken:write(v .. '\n')
          frefresh:write(v .. '|' .. refresh_tokens[i] .. '\n')
        end
    end
    ftoken:flush()
    ftoken:close()
    frefresh:flush()
    frefresh:close()
    local msg = "req_count: %d, error_total: %d, errors: %s"
    print(msg:format(req_total, error_total, format(errors_all)))
end

-- 下面设置请求方法，请求参数，请求头
-- wrk.body = "{\"activity_id\":102000022,\"items\":[{\"sku_id\":833,\"count\":1}]}"
wrk.method = 'POST'
wrk.headers['content-type'] = 'application/json'
