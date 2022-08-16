import * as dagreD3 from 'dagre-d3'
import * as d3 from 'd3'
interface CheckboxModel {
    [key: string]: string
}
export default class Dag {
    g: any;
    node_shape = "rect";
    node_stype = "fill:#fff;stroke:#000;";
    edge_stype = "fill:#fff;stroke:#333;stroke-width:1.5px";
    info_bg = "#000";
    info_font = "#fff";
    info_font_size = "0.7em"
    rankdir = "TB";
    tooltip_css = ''
    dsl: any
    conf: any
    canva_selector: any
    info_selector: any
    d3 = d3 as any
    checkboxArray: any[] = []
    checkboxFirst = ''
    constructor(dsl: string, conf: any, canva_selector: any, info_selector: any = null) {
        this.dsl = typeof dsl == 'string' ? JSON.parse(dsl) : dsl;
        this.conf = typeof conf == 'string' ? JSON.parse(conf) : conf;
        this.canva_selector = canva_selector
        this.info_selector = info_selector
    }

    get_source_data(input_filed: any): any {
        var source_inputs = new Array();
        for (var key in input_filed) {
            if (!Array.isArray(input_filed[key])) {
                return this.get_source_data(input_filed[key]);
            } else {
                input_filed[key].forEach((val: string) => {
                    source_inputs.push(val);
                });
            }
        }
        return source_inputs
    }

    handle_common_components(component_name: string, common_components: any, results: any) {
        var iter = function (obj: any, results: any[]) {
            for (var key in obj) {
                if (key == component_name) {
                    results.push(obj[key])
                    return results;
                } else {
                    if (typeof obj[key] == "object") {
                        iter(obj[key], results);
                    } else {
                        return results;
                    }
                }
            }

            return results
        }

        return iter(common_components, results)
    }

    handle_role_components(component_name: string, role_components: any, results: any) {
        var found = false
        var iter = function (obj: any, results: any[]) {
            for (var key in obj) {
                if (key == component_name) {
                    found = true
                    results.push(obj[key])
                    return results;
                } else {
                    if (key == "guest" || key == "host") {
                        results.push({
                            "role": key
                        })
                    }

                    if (typeof obj[key] == "object") {
                        iter(obj[key], results);
                    } else {
                        return results;
                    }
                }
            }
            if (!found) {
                results.pop()
            }

            return results
        }

        return iter(role_components, results)
    }

    get_component_para(component_name: string) {
        var components_paras = this.conf["component_parameters"];
        var results = new Array();
        if (components_paras.hasOwnProperty("common")) {
            this.handle_common_components(component_name, components_paras["common"], results)
        }

        if (results.length == 0 && components_paras.hasOwnProperty("role")) {
            this.handle_role_components(component_name, components_paras["role"], results)
        }

        return results
    }

    reset_canvas(canvas: any) {
        canvas.selectAll("*").remove();
    }

    print_parameter(parameters: any) {
        var parameters_str = ''
        if (this.tooltip_css === 'dagTips') {
            parameters_str = `
          <div style='position: absolute;right: -260px;top:0; height: 400px; padding-top: 4px'><table style=" border-right: 1px solid #333; border-bottom: 1px solid #333;">
        <tbody>
          <tr style='background: #D8E3E9'>
            <th style='border-top: 1px solid #333; border-left: 1px solid #333;'>key</th>
            <th style='border-top: 1px solid #333; border-left: 1px solid #333;'>value</th>
          </tr>
          <tr>`
        } else {
            parameters_str = `<div style='position: absolute;right: 10px;top:0; height: 450px; padding-top: 4px'><table style=" border-right: 1px solid #333; border-bottom: 1px solid #333; background: #fff;">
          <tbody>
            <tr style='background: #D8E3E9'>
              <th style='border-top: 1px solid #333; border-left: 1px solid #333;'>key</th>
              <th style='border-top: 1px solid #333; border-left: 1px solid #333;'>value</th>
            </tr>
            <tr>`
        }
        parameters.forEach((onepara: any) => {
            var str = ''
            for (var key in onepara) {
                if (key == "role") {
                    str += `<td style='border-top: 1px solid #333; border-left: 1px solid #333;'>Role</td>
                    <td style='border-top: 1px solid #333; border-left: 1px solid #333;'>${onepara[key]}</td>`
                    str += "</tr>"
                } else {
                    str += '<tr>'
                    str += "<td style='border-top: 1px solid #333; border-left: 1px solid #333;'>" + key + "</td>"
                    str += "<td style='width: 200px;word-break: break-word; border-top: 1px solid #333; border-left: 1px solid #333;'>" + JSON.stringify(onepara[key]) + "</td>"
                    str += '</tr>'
                }
            }
            parameters_str += str
        });
        return parameters_str + '</tbody></table></div>';
    }

    Generate() {
        this.g = new dagreD3.graphlib.Graph();
        this.g.setGraph({
            rankdir: this.rankdir
        });
        for (let component_name in this.dsl.components) {
            var one_model = this.dsl.components[component_name]
            this.g.setNode(component_name, {
                label: component_name,
                rx: 5,
                ry: 5,
                shape: this.node_shape,
                style: this.node_stype
            });
        }
        for (let component_name in this.dsl.components) {
            var one_model = this.dsl.components[component_name];
            if (one_model.hasOwnProperty("input")) {
                var inputs = this.get_source_data(one_model["input"]);
                this.getComponentRelation(component_name, inputs[0].split(".")[0])
                inputs.forEach((one_input: string) => {
                    var source_component = one_input.split(".")[0];
                    this.g.setEdge(source_component, component_name, {
                        label: "",
                        style: this.edge_stype
                    });
                });
            } else {
                this.checkboxFirst = component_name
            }
        }
        this.my[0] = {
            name: this.checkboxFirst,
            value: true
        }
        this.mapCheckbox(this.checkboxFirst)
    }

    Draw() {
        let render = new dagreD3.render();
        let svg = d3.select(this.canva_selector);
        this.reset_canvas(svg)
        let svgGroup: any = svg.append('g');
        render(svgGroup, this.g);
        if (this.info_selector == null) {
            let info = d3.select("body").append("div")
                .style("position", "absolute")
                .style("opacity", 0)
                .style("background-color", this.info_bg)
                .style("font-size", this.info_font_size)
                .style("color", this.info_font)
                .style("padding-left", "10px")
                .style("padding-right", "10px")
                .style("display", "none");

            // mouseover for detail
            svg.selectAll("g.node").on("mouseenter", (e: any) => {
                info.transition()
                    .duration(400)
                    .style('opacity', 0.9)
                    .style('display', 'block')
                    .style('position', 'absolute');
                info.html(this.print_parameter(this.get_component_para(e)) + '</div>')
                    .style('left', (this.d3.event.pageX + 35) + 'px')
                    .style('top', (this.d3.event.pageY + 5) + 'px');
            })
                .on("mouseleave", e => {
                    info.transition()
                        .duration(400)
                        .style('opacity', 0)
                        .style('display', 'none');
                });
        } else {
            let info = d3.select(this.info_selector);
            svg.selectAll("g.node").on("mouseenter", (e: any) => {
                info.html(this.print_parameter(this.get_component_para(e)));
            })
                .on("mouseleave", e => {
                    info.html("");
                });

        }

        svg.selectAll("g.edgePath").on("mouseenter", e => {
        });

        svg.attr("preserveAspectRatio", "xMidYMin").attr("viewBox", "0 0 800 1100")

    }
    count = 0
    my: any[] = []
    cObj: any = {}
    getComponentRelation(child: string, component_name: string) {
        if (this.cObj.hasOwnProperty(component_name)) {
            this.cObj[component_name].children.push(child)
        } else {
            this.cObj[component_name] = {
                name: component_name,
                children: [child]
            }
        }
    }
    mapCheckbox(name: string): any {
        if (!this.cObj[name]) {
            if (this.find(name)) {
                this.my.push({
                    name: name,
                    value: false
                })
            }
            this.checkboxFirst = name
            return false
        } else if (this.cObj[name].children.length < 2) {
            if (this.find(this.cObj[name].children[0])) {
                if (this.cObj[name].children[0].indexOf('evaluation') !== -1 || this.cObj[name].children[0].indexOf('homo_data_split_0') !== -1) {
                    this.my.push({
                        name: this.cObj[name].children[0],
                        value: false
                    })
                } else {
                    this.my.push({
                        name: this.cObj[name].children[0],
                        value: true
                    })
                }
            }
            this.checkboxFirst = this.cObj[name].children[0]
            return this.mapCheckbox(this.checkboxFirst)
        } else if (this.cObj[name].children.length > 2) {
            this.cObj[name].children.forEach((el: string) => {
                if (this.find(el)) {
                    if (el.indexOf('evaluation') !== -1  || el.indexOf('homo_data_split_0') !== -1) {
                        this.my.push({
                            name: el,
                            value: false
                        })
                    } else {
                        this.my.push({
                            name: el,
                            value: false
                        })
                    }
                }
            });
            this.cObj[name].children.forEach((el: string) => {
                this.checkboxFirst = el
                return this.mapCheckbox(this.checkboxFirst)
            });
        }
    }
    find(item: any) {
        return this.my.findIndex(c => c.name === item) === -1 ? true : false
    }
}