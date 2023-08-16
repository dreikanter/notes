class Notes::CleanshotDownloader
  attr_reader :cleanshot_url

  def initialize(cleanshot_url, file_name: nil)
    @cleanshot_url = cleanshot_url
    @file_name = file_name
  end

  def download
    response = HTTP.get(direct_image_url)
    raise "error downloading image" unless response.status.success?
    File.open(file_name, "wb") { _1.write(response.body.to_s) }
    file_name
  end

  private

  def direct_image_url
    @direct_image_url ||= begin
      response = HTTP.get("#{cleanshot_url}+")
      raise "error getting image URL" unless response.status.redirect?
      response.headers["Location"]
    end
  end

  def file_name
    @file_name ||= begin
      query = URI.parse(direct_image_url).query
      param = CGI.parse(query).fetch("response-content-disposition").first
      CGI.unescape(param.split("=", 2).last)
    end
  end
end
