class Notes::CleanshotDownloader
  attr_reader :url

  def initialize(url)
    @url = url
  end

  def download
    response = HTTP.get(direct_image_url)
    raise "error downloading image" unless response.status.success?
    { original_file_name: original_file_name, content: response.body.to_s }
  end

  private

  def direct_image_url
    @direct_image_url ||= begin
      response = HTTP.get("#{url}+")
      raise "error getting image URL" unless response.status.redirect?
      response.headers["Location"]
    end
  end

  def original_file_name
    query = URI.parse(direct_image_url).query
    param = CGI.parse(query).fetch("response-content-disposition").first
    CGI.unescape(param.split("=", 2).last)
  end
end
