use std::fs;

fn read_load_avg() -> Option<f64> {
    let contents = fs::read_to_string("/proc/loadavg").ok()?;
    let first_field = contents.split_whitespace().next()?;
    first_field.parse::<f64>().ok()
}

fn read_mem_info() -> Option<(u64, u64)> {
    let contents = fs::read_to_string("/proc/meminfo").ok()?;
    let mut total: Option<u64> = None;
    let mut available: Option<u64> = None;

    for line in contents.lines() {
        if let Some(rest) = line.strip_prefix("MemTotal:") {
            total = parse_kb_field(rest);
        } else if let Some(rest) = line.strip_prefix("MemAvailable:") {
            available = parse_kb_field(rest);
        }
    }

    match (total, available) {
        (Some(t), Some(a)) => Some((t, a)),
        _ => None,
    }
}

fn parse_kb_field(field: &str) -> Option<u64> {
    field.split_whitespace().next()?.parse::<u64>().ok()
}

fn main() {
    let load_avg = read_load_avg();
    let mem = read_mem_info();

    if load_avg.is_none() && mem.is_none() {
        println!("{{\"error\":\"proc unavailable\"}}");
        return;
    }

    let mut fields: Vec<String> = Vec::new();

    if let Some(load) = load_avg {
        fields.push(format!("\"load_avg_1m\":{:.2}", load));
    }

    if let Some((total_kb, available_kb)) = mem {
        let total_mb = total_kb / 1024;
        let available_mb = available_kb / 1024;
        let used_percent = if total_kb > 0 {
            (total_kb - available_kb) as f64 / total_kb as f64 * 100.0
        } else {
            0.0
        };
        fields.push(format!("\"mem_total_mb\":{}", total_mb));
        fields.push(format!("\"mem_available_mb\":{}", available_mb));
        fields.push(format!("\"mem_used_percent\":{:.1}", used_percent));
    }

    println!("{{{}}}", fields.join(","));
}
