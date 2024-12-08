use clap::Parser;
use serde::{Deserialize, Serialize};
use serde_json;
use serde_sexpr;
use std::fs;

type Spec = Vec<Element>;

#[derive(Serialize, Deserialize)]
enum Element {
    Workdir(String),
    User(String, String),
    Volume(Vec<String>),
    Expose(u16, String),
    Labels(Vec<(String, String)>),
    Run(Vec<String>),
    From(String),
    Env(String, String),
    Copy(String, String),
    Cmd(Vec<String>),
    Add(String, String),
}

// WORKDIR /path/to/workdir
fn oci_workdir(path: String) -> String {
    format!("WORKDIR {}\n", path)
}

// USER <user>:<group>
fn oci_user(user: String, group: String) -> String {
    format!("USER {user}:{group}\n")
}

// VOLUME ["/data"]
fn oci_volume(vols: Vec<String>) -> String {
    let v: Vec<String> = vols.into_iter().map(|s| format!("\"{s}\"")).collect();
    let s = v.join(",");
    format!("VOLUME [{s}]\n")
}

// EXPOSE <port>/<protocol>
fn oci_expose(port: u16, proto: String) -> String {
    format!("EXPOSE {port}/{proto}\n")
}

// LABEL <key>=<value> <key>=<value> <key>=<value> ...
fn oci_label(labels: Vec<(String, String)>) -> String {
    let v: Vec<String> = labels
        .into_iter()
        .map(|(key, val)| format!("{key}={val}"))
        .collect();
    let s = v.join(" ");
    format!("LABEL {s}\n")
}

// RUN [ "<command>", ... ]
fn oci_run(args: Vec<String>) -> String {
    let v: Vec<String> = args.into_iter().map(|s| format!("\"{s}\"")).collect();
    let s = v.join(",");
    format!("RUN [{s}]\n")
}

// FROM <base>
fn oci_from(base: String) -> String {
    format!("FROM {base}\n")
}

// ENV <key>=<value> ...
fn oci_env(key: String, val: String) -> String {
    format!("ENV {key}={val}\n")
}

// COPY <src> <dest>
fn oci_copy(src: String, dest: String) -> String {
    format!("COPY {src} {dest}\n")
}

// CMD ["executable", "param1", "param2"]
fn oci_cmd(args: Vec<String>) -> String {
    let v: Vec<String> = args.into_iter().map(|s| format!("\"{s}\"")).collect();
    let s = v.join(",");
    format!("CMD [{s}]\n")
}

// ADD <src> <dest>
fn oci_add(src: String, dest: String) -> String {
    format!("ADD {src} {dest}\n")
}

fn mkoci(spec: Spec) -> String {
    let mut finals: String = String::new();
    for elem in spec {
        match elem {
            Element::Workdir(wd) => finals.push_str(&oci_workdir(wd)),
            Element::User(u, g) => finals.push_str(&oci_user(u, g)),
            Element::Volume(v) => finals.push_str(&oci_volume(v)),
            Element::Expose(po, pr) => finals.push_str(&oci_expose(po, pr)),
            Element::Labels(l) => finals.push_str(&oci_label(l)),
            Element::Run(a) => finals.push_str(&oci_run(a)),
            Element::From(b) => finals.push_str(&oci_from(b)),
            Element::Env(k, v) => finals.push_str(&oci_env(k, v)),
            Element::Copy(s, d) => finals.push_str(&oci_copy(s, d)),
            Element::Cmd(a) => finals.push_str(&oci_cmd(a)),
            Element::Add(s, d) => finals.push_str(&oci_add(s, d)),
        }
    }
    finals
}

// Application section

#[derive(clap::ValueEnum, Debug, Clone)]
enum ParseMode {
    Json,
    Lisp,
    Ron,
}

#[derive(Parser, Debug)]
#[clap(version, about = "Make Dockerfile from known format")]
struct Arguments {
    mode: ParseMode,
    file: String,
}

fn mkrev() -> String {
    let mut nspec = Spec::new();
    nspec.push(Element::From("scratch".to_owned()));
    nspec.push(Element::Add("data.tar.gz".to_owned(), "/".to_owned()));
    nspec.push(Element::Run(vec!["/usr/bin/ldconfig".to_owned()]));
    nspec.push(Element::Workdir("/root".to_owned()));
    nspec.push(Element::Cmd(vec!["/bin/bash".to_owned(), "-i".to_owned()]));
    return serde_sexpr::to_string(&nspec).unwrap();
}

fn main() {
    let args = Arguments::parse();
    let content = fs::read_to_string(args.file).unwrap();

    let ispec: Spec = match args.mode {
        ParseMode::Json => serde_json::from_str(&content).unwrap(),
        ParseMode::Lisp => serde_sexpr::from_str(&content).unwrap(),
        ParseMode::Ron => ron::from_str(&content).unwrap(),
    };

    print!("{}", mkoci(ispec))
}
